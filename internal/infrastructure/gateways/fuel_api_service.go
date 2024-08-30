package gateways

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

//go:generate mockgen -source fuel_api_service.go -destination mocks/fuel_api_service_mock.go -package mocks
type FuelAPIService interface {
	FetchDeviceImages(mk, mdl string, yr int, prodID int, prodFormat int) (FuelDeviceImages, error)
}

// fuelAPIService client
type fuelAPIService struct {
	vehicleURL url.URL
	imageURL   url.URL
	key        string
	log        *zerolog.Logger
}

func NewFuelAPIService(settings *config.Settings, logger *zerolog.Logger) FuelAPIService {
	fa := &fuelAPIService{
		vehicleURL: settings.FuelAPIVehiclesEndpoint,
		imageURL:   settings.FuelAPIImagesEndpoint,
		key:        settings.FuelAPIKey,
		log:        logger,
	}
	if fa.vehicleURL.String() == "" {
		logger.Fatal().Msgf("fuel api vehicle url is empty")
	}
	if fa.imageURL.String() == "" {
		logger.Fatal().Msgf("fuel api image url is empty")
	}
	if fa.key == "" {
		logger.Fatal().Msgf("fuel api key is empty")
	}
	return fa
}

func (fs *fuelAPIService) FetchDeviceImages(mk, mdl string, yr int, prodID int, prodFormat int) (FuelDeviceImages, error) {
	notExactImage := false // if we pull image where year doesn't match
	// search for exact MMY image
	img, err := fs.imageRequest(mk, mdl, yr, prodID, prodFormat)
	if err != nil {
		return FuelDeviceImages{}, err
	}

	// search for model and make (remove year)
	if !img.validURL {
		img, err = fs.imageRequest(mk, mdl, 0, prodID, prodFormat)
		if err != nil {
			return FuelDeviceImages{}, err
		}
		notExactImage = true

		// search for model and first work of make
		// ex: Wrangler Sport -> Wrangler
		if !img.validURL {
			m := strings.Split(mdl, " ")
			img, err = fs.imageRequest(mk, m[0], 0, prodID, prodFormat)
			if err != nil {
				return FuelDeviceImages{}, err
			}
		}
	}

	if !img.validURL {
		fs.log.Log().Msgf("request for device image unsuccessful: %s %s %d", mk, mdl, yr)
		return FuelDeviceImages{}, errors.New("request for device image unsuccessful")
	}
	img.NotExactImage = notExactImage

	return img, nil

}

func (fs *fuelAPIService) imageRequest(mk, mdl string, yr int, prodID int, prodFormat int) (FuelDeviceImages, error) {
	vehicleReqURL := fmt.Sprintf("?year=%d&make=%s&model=%s&api_key=%s", yr, mk, mdl, fs.key)
	vehicleResp, err := http.Get(fs.vehicleURL.String() + vehicleReqURL)
	if err != nil {
		return FuelDeviceImages{}, err
	}
	if vehicleResp.StatusCode >= 400 {
		fs.log.Info().Msgf("bad request status: %d", vehicleResp.StatusCode)
		return FuelDeviceImages{}, errors.New("unable to fetch vehicle data: bad requset")
	}

	vehicleData, err := io.ReadAll(vehicleResp.Body)
	if err != nil {
		return FuelDeviceImages{}, err
	}
	vehicleID := gjson.Get(string(vehicleData), "0.id").Str
	imageReqURL := fmt.Sprintf("/%s?api_key=%s&productID=%d&productFormatIDs=%d", vehicleID, fs.key, prodID, prodFormat)
	imageResp, err := http.Get(fs.imageURL.String() + imageReqURL)
	if err != nil {
		return FuelDeviceImages{}, err
	}
	if imageResp.StatusCode >= 400 {
		fs.log.Info().Msgf("bad request status: %d", imageResp.StatusCode)
		return FuelDeviceImages{}, errors.New("unable to fetch image: bad requset")
	}

	response, err := io.ReadAll(imageResp.Body)
	if err != nil {
		return FuelDeviceImages{}, err
	}
	imageData := string(response)

	width := gjson.Get(imageData, "products.0.productFormats.0.width").Int()
	height := gjson.Get(imageData, "products.0.productFormats.0.height").Int()
	angle := gjson.Get(imageData, "products.0.productFormats.0.angle").String()
	img := FuelDeviceImages{FuelAPIID: vehicleID, Width: int(width), Height: int(height), Angle: angle, Images: make([]FuelImage, 0)}
	gjson.Get(imageData, "products.0.productFormats.0.assets").ForEach(func(_ gjson.Result, value gjson.Result) bool {
		imageURL := value.Get("url").Str
		color := value.Get("shotCode.color.simple_name").Str
		img.Images = append(img.Images, FuelImage{SourceURL: imageURL, Color: color})
		if !img.validURL && imageURL != "" {
			img.validURL = true
		}
		return true
	})

	return img, nil

}

type FuelDeviceImages struct {
	FuelAPIID string      `boil:"fuelID"`
	Width     int         `boil:"width"`
	Height    int         `boil:"height"`
	Angle     string      `boil:"angle"`
	Images    []FuelImage `boil:"images"`
	validURL  bool
	// use to track if we used a different year image and could not find this one
	NotExactImage bool
}

type FuelImage struct {
	SourceURL string `boil:"sourceURL"`
	Color     string `boil:"color"`
}
