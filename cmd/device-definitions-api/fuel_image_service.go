package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

// ElasticSearchService client
type FuelServiceAPI struct {
	VehicleURL string
	ImageURL   string
	Key        string
	log        *zerolog.Logger
	db         *db.Store
}

type deviceData struct {
	Make  string
	Model string
	Year  int
}

func fetchFuelAPIImages(ctx context.Context, logger zerolog.Logger, settings *config.Settings) error {

	fs := NewFuelService(ctx, settings, &logger)
	fs.fetchDeviceImage("tesla", "model-3", 2019)
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
	}
}

func (fs *FuelServiceAPI) fetchDeviceImage(mk, mdl string, yr int) (string, error) {
	vehicleReqURL := fmt.Sprintf("?year=%d&make=%s&model=%s&api_key=%s", yr, mk, mdl, fs.Key)

	fmt.Println(fs.VehicleURL + vehicleReqURL)
	vehicleResp, err := http.Get(fs.VehicleURL + vehicleReqURL)
	if err != nil {
		return "", err
	}
	vehicleData, err := ioutil.ReadAll(vehicleResp.Body)
	if err != nil {
		return "", err
	}
	vehicleID := gjson.Get(string(vehicleData), "0.id").Str
	imageReqURL := fmt.Sprintf("/%s?api_key=%s", vehicleID, fs.Key)
	imageResp, err := http.Get(fs.ImageURL + imageReqURL)
	if err != nil {
		return "", err
	}
	imageData, err := ioutil.ReadAll(imageResp.Body)
	if err != nil {
		return "", err
	}
	imageURL := gjson.Get(string(imageData), "products.0.productFormats.0.assets.0.url").Str

	return imageURL, nil

}

func (fs *FuelServiceAPI) deviceData(ctx context.Context) ([]deviceData, error) {

	oems, err := models.DeviceMakes().All(ctx, fs.db.DBS().Reader)
	if err != nil {
		return []deviceData{}, err
	}

	for _, mk := range oems {
		mdls, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(mk.ID)).All(ctx, fs.db.DBS().Reader)
		if err != nil {
			return []deviceData{}, err
		}

		for _, mdl := range mdls {
			fmt.Println(mdl)
		}
	}

	return []deviceData{}, nil

}

// SELECT m.id, m.name, m.name_slug, d.device_make_id, d.id, d.model_slug, d.year
// FROM device_definitions_api.device_makes m
// LEFT JOIN (
//     SELECT device_make_id, id, model_slug, year
//     FROM device_definitions_api.device_definitions_api.device_definitions
// ) d
// ON d.device_make_id = m.id
// LIMIT 50;
