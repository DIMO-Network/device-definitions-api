package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// ElasticSearchService client
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

type image struct {
	FuelAPIID string `boil:"fuel_api_id"`
	Width     int    `boil:"width"`
	Height    int    `boil:"height"`
	Angle     string `boil:"angle"`
	SourceURL string `boil:"source_url"`
	Color     string `boil:"color"`
}

func fetchFuelAPIImages(ctx context.Context, logger zerolog.Logger, settings *config.Settings) error {

	fs := NewFuelService(ctx, settings, &logger)
	devices, err := fs.deviceData(ctx)
	if err != nil {
		return err
	}
	err = fs.writeToTable(ctx, devices, 2, 2)
	if err != nil {
		fmt.Println(err)
	}
	err = fs.writeToTable(ctx, devices, 2, 6)
	if err != nil {
		fmt.Println(err)
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

			img, err := fs.fetchDeviceImage(d.Make, d.Models[n].Model, d.Models[n].Year, prodID, prodFormat)
			if err != nil {
				continue
			}

			var p models.Image
			p.ID = ksuid.New().String()
			p.DeviceDefinitionID = d.Models[n].DeviceDefinitionID
			p.FuelAPIID = null.StringFrom(img.FuelAPIID)
			p.Width = null.IntFrom(img.Width)
			p.Height = null.IntFrom(img.Height)
			p.SourceURL = img.SourceURL
			p.DimoS3URL = null.StringFrom("")
			p.Color = img.Color

			err = p.Upsert(ctx, fs.db.Writer, false, []string{models.ImageColumns.DeviceDefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fs *FuelServiceAPI) fetchDeviceImage(mk, mdl string, yr int, prodID int, prodFormat int) (image, error) {

	img, err := fs.imageRequest(mk, mdl, yr, prodID, prodFormat)
	if err != nil {
		return image{}, err
	}

	if img.SourceURL == "" {
		img, err = fs.imageRequest(mk, mdl, 0, prodID, prodFormat)
		if err != nil {
			return image{}, err
		}

		if img.SourceURL == "" {
			m := strings.Split(mdl, " ")
			img, err = fs.imageRequest(mk, m[0], 0, prodID, prodFormat)
			if err != nil {
				return image{}, err
			}
		}
	}

	if img.SourceURL == "" {
		fs.log.Log().Msgf("request for device image unsuccessful: %s %s %d", mk, mdl, yr)
		return image{}, errors.New("request for device image unsuccessful")
	}

	return img, nil

}

func (fs *FuelServiceAPI) imageRequest(mk, mdl string, yr int, prodID int, prodFormat int) (image, error) {
	vehicleReqURL := fmt.Sprintf("?year=%d&make=%s&model=%s&api_key=%s", yr, mk, mdl, fs.Key)
	vehicleResp, err := http.Get(fs.VehicleURL + vehicleReqURL)
	if err != nil {
		return image{}, err
	}
	vehicleData, err := ioutil.ReadAll(vehicleResp.Body)
	if err != nil {
		return image{}, err
	}
	vehicleID := gjson.Get(string(vehicleData), "0.id").Str
	imageReqURL := fmt.Sprintf("/%s?api_key=%s&productID=%d&productFormatIDs=%d", vehicleID, fs.Key, prodID, prodFormat)
	imageResp, err := http.Get(fs.ImageURL + imageReqURL)
	if err != nil {
		return image{}, err
	}
	imageData, err := ioutil.ReadAll(imageResp.Body)
	if err != nil {
		return image{}, err
	}
	imageURL := gjson.Get(string(imageData), "products.0.productFormats.0.assets.0.url").Str
	width := gjson.Get(string(imageData), "products.0.productFormats.0.width").Int()
	height := gjson.Get(string(imageData), "products.0.productFormats.0.height").Int()
	angle := gjson.Get(string(imageData), "products.0.productFormats.0.angle").String()
	color := gjson.Get(string(imageData), "products.0.productFormats.0.assets.0.shotCode.color.simple_name").Str
	img := image{FuelAPIID: vehicleID, Width: int(width), Height: int(height), Angle: angle, SourceURL: imageURL, Color: color}
	return img, nil

}

func (fs *FuelServiceAPI) deviceData(ctx context.Context) ([]deviceData, error) {

	oems, err := models.DeviceMakes().All(ctx, fs.db.Reader)
	if err != nil {
		return []deviceData{}, err
	}

	devices := make([]deviceData, len(oems))
	for n, mk := range oems {
		mdls, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(mk.ID)).All(ctx, fs.db.Reader)
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
