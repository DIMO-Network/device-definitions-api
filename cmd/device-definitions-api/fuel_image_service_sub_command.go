package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type syncFuelImageCmd struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*syncFuelImageCmd) Name() string     { return "images" }
func (*syncFuelImageCmd) Synopsis() string { return "images args to stdout." }
func (*syncFuelImageCmd) Usage() string {
	return `images [] <some text>:
	images args.
  `
}

func (p *syncFuelImageCmd) SetFlags(f *flag.FlagSet) {

}

func (p *syncFuelImageCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := fetchFuelAPIImages(ctx, p.logger, &p.settings)
	if err != nil {
		p.logger.Error().Err(err)
	}
	return subcommands.ExitSuccess
}

// FuelServiceAPI client
type FuelServiceAPI struct {
	VehicleURL string
	ImageURL   string
	Key        string
	log        *zerolog.Logger
	db         *db.ReaderWriter
}

type deviceData struct {
	Make   string
	Models []model
}

type model struct {
	Model              string
	Year               int
	DeviceDefinitionID string
}

type deviceImages struct {
	FuelAPIID string  `boil:"fuelID"`
	Width     int     `boil:"width"`
	Height    int     `boil:"height"`
	Angle     string  `boil:"angle"`
	Images    []image `boil:"images"`
	validURL  bool
	// use to track if we used a different year image and could not find this one
	NotExactImage bool
}

type image struct {
	SourceURL string `boil:"sourceURL"`
	Color     string `boil:"color"`
}

func fetchFuelAPIImages(ctx context.Context, logger zerolog.Logger, settings *config.Settings) error {

	fs := NewFuelService(ctx, settings, &logger)
	devices, err := fs.deviceData(ctx)
	if err != nil {
		return err
	}
	logger.Info().Msgf("pulling fuel images for %d device definitions", len(devices))

	err = fs.writeToTable(ctx, devices, 2, 2)
	if err != nil {
		logger.Err(err).Msg("failed to writeToTable when fetching Fuel API images")
	}
	err = fs.writeToTable(ctx, devices, 2, 6)
	if err != nil {
		logger.Err(err).Msg("failed to writeToTable when fetching Fuel API images")
	}

	return nil
}

func NewFuelService(ctx context.Context, settings *config.Settings, logger *zerolog.Logger) *FuelServiceAPI {

	pdb := db.NewDbConnectionFromSettings(ctx, &settings.DB, true)
	pdb.WaitForDB(*logger)

	return &FuelServiceAPI{
		VehicleURL: settings.FuelAPIVehiclesEndpoint,
		ImageURL:   settings.FuelAPIImagesEndpoint,
		Key:        settings.FuelAPIKey,
		log:        logger,
		db:         pdb.DBS(),
	}
}

func (fs *FuelServiceAPI) writeToTable(ctx context.Context, data []deviceData, prodID int, prodFormat int) error {

	for _, d := range data {
		for n := range d.Models {
			img, err := fs.fetchDeviceImages(d.Make, d.Models[n].Model, d.Models[n].Year, prodID, prodFormat)
			if err != nil {
				fs.log.Warn().Msgf("unable to fetch device image for: %d %s %s", d.Models[n].Year, d.Make, d.Models[n].Model)
				continue
			}
			var p models.Image

			// loop through all img (color variations)
			for _, device := range img.Images {
				p.ID = ksuid.New().String()
				p.DeviceDefinitionID = d.Models[n].DeviceDefinitionID
				p.FuelAPIID = null.StringFrom(img.FuelAPIID)
				p.Width = null.IntFrom(img.Width)
				p.Height = null.IntFrom(img.Height)
				p.SourceURL = device.SourceURL
				//p.DimoS3URL = null.StringFrom("") // dont set it so it is null
				p.Color = device.Color
				p.NotExactImage = img.NotExactImage

				err = p.Upsert(ctx, fs.db.Writer, true, []string{models.ImageColumns.DeviceDefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (fs *FuelServiceAPI) fetchDeviceImages(mk, mdl string, yr int, prodID int, prodFormat int) (deviceImages, error) {
	notExactImage := false // if we pull image where year doesn't match
	// search for exact MMY image
	img, err := fs.imageRequest(mk, mdl, yr, prodID, prodFormat)
	if err != nil {
		return deviceImages{}, err
	}

	// search for model and make (remove year)
	if !img.validURL {
		img, err = fs.imageRequest(mk, mdl, 0, prodID, prodFormat)
		if err != nil {
			return deviceImages{}, err
		}
		notExactImage = true

		// search for model and first work of make
		// ex: Wrangler Sport -> Wrangler
		if !img.validURL {
			m := strings.Split(mdl, " ")
			img, err = fs.imageRequest(mk, m[0], 0, prodID, prodFormat)
			if err != nil {
				return deviceImages{}, err
			}
		}
	}

	if !img.validURL {
		fs.log.Log().Msgf("request for device image unsuccessful: %s %s %d", mk, mdl, yr)
		return deviceImages{}, errors.New("request for device image unsuccessful")
	}
	img.NotExactImage = notExactImage

	return img, nil

}

func (fs *FuelServiceAPI) imageRequest(mk, mdl string, yr int, prodID int, prodFormat int) (deviceImages, error) {
	vehicleReqURL := fmt.Sprintf("?year=%d&make=%s&model=%s&api_key=%s", yr, mk, mdl, fs.Key)
	vehicleResp, err := http.Get(fs.VehicleURL + vehicleReqURL)
	if err != nil {
		return deviceImages{}, err
	}
	if vehicleResp.StatusCode >= 400 {
		fs.log.Info().Msgf("bad request status: %d", vehicleResp.StatusCode)
		return deviceImages{}, errors.New("unable to fetch vehicle data: bad requset")
	}

	vehicleData, err := io.ReadAll(vehicleResp.Body)
	if err != nil {
		return deviceImages{}, err
	}
	vehicleID := gjson.Get(string(vehicleData), "0.id").Str
	imageReqURL := fmt.Sprintf("/%s?api_key=%s&productID=%d&productFormatIDs=%d", vehicleID, fs.Key, prodID, prodFormat)
	imageResp, err := http.Get(fs.ImageURL + imageReqURL)
	if err != nil {
		return deviceImages{}, err
	}
	if imageResp.StatusCode >= 400 {
		fs.log.Info().Msgf("bad request status: %d", imageResp.StatusCode)
		return deviceImages{}, errors.New("unable to fetch image: bad requset")
	}

	response, err := io.ReadAll(imageResp.Body)
	if err != nil {
		return deviceImages{}, err
	}
	imageData := string(response)

	width := gjson.Get(imageData, "products.0.productFormats.0.width").Int()
	height := gjson.Get(imageData, "products.0.productFormats.0.height").Int()
	angle := gjson.Get(imageData, "products.0.productFormats.0.angle").String()
	img := deviceImages{FuelAPIID: vehicleID, Width: int(width), Height: int(height), Angle: angle, Images: make([]image, 0)}
	gjson.Get(imageData, "products.0.productFormats.0.assets").ForEach(func(key gjson.Result, value gjson.Result) bool {
		imageURL := value.Get("url").Str
		color := value.Get("shotCode.color.simple_name").Str
		img.Images = append(img.Images, image{SourceURL: imageURL, Color: color})
		if !img.validURL && imageURL != "" {
			img.validURL = true
		}
		return true
	})

	return img, nil

}

// deviceData looks for makes and models in our database and returns a projection of them specific to Fuel
func (fs *FuelServiceAPI) deviceData(ctx context.Context) ([]deviceData, error) {

	oems, err := models.DeviceMakes().All(ctx, fs.db.Reader)
	if err != nil {
		return []deviceData{}, err
	}

	devices := make([]deviceData, len(oems))
	for n, mk := range oems {
		mdls, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(mk.ID),
			models.DeviceDefinitionWhere.Year.GTE(2005)).All(ctx, fs.db.Reader)
		if err != nil {
			return []deviceData{}, err
		}
		devices[n] = deviceData{Make: mk.NameSlug, Models: make([]model, len(mdls))}
		for i, mdl := range mdls {
			devices[n].Models[i] = model{Model: mdl.ModelSlug, Year: int(mdl.Year), DeviceDefinitionID: mdl.ID}
		}
	}

	return devices, nil

}
