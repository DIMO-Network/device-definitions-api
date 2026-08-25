//nolint:tagliatelle
package config

import (
	"net/url"

	"github.com/DIMO-Network/shared/pkg/db"

	"github.com/DIMO-Network/shared/pkg/redis"
)

type Settings struct {
	Environment                       string         `yaml:"ENVIRONMENT"`
	Port                              string         `yaml:"PORT"`
	MonitoringPort                    string         `yaml:"MONITORING_PORT"`
	LogLevel                          string         `yaml:"LOG_LEVEL"`
	DB                                db.Settings    `yaml:"DB"`
	ServiceName                       string         `yaml:"SERVICE_NAME"`
	ServiceVersion                    string         `yaml:"SERVICE_VERSION"`
	GRPCPort                          string         `yaml:"GRPC_PORT"`
	IPFSNodeEndpoint                  string         `yaml:"IPFS_NODE_ENDPOINT"`
	DrivlyAPIKey                      string         `yaml:"DRIVLY_API_KEY"`
	DrivlyVINAPIURL                   url.URL        `yaml:"DRIVLY_VIN_API_URL"`
	DrivlyOfferAPIURL                 url.URL        `yaml:"DRIVLY_OFFER_API_URL"`
	Redis                             redis.Settings `yaml:"REDIS"`
	FuelAPIVehiclesEndpoint           url.URL        `yaml:"FUEL_API_VEHICLES_ENDPOINT"`
	FuelAPIImagesEndpoint             url.URL        `yaml:"FUEL_API_IMAGES_ENDPOINT"`
	FuelAPIKey                        string         `yaml:"FUEL_API_KEY"`
	VincarioAPIURL                    url.URL        `yaml:"VINCARIO_API_URL"`
	VincarioAPIKey                    string         `yaml:"VINCARIO_API_KEY"`
	VincarioAPISecret                 string         `yaml:"VINCARIO_API_SECRET"`
	AutoIsoAPIUid                     string         `yaml:"AUTO_ISO_API_UID"`
	AutoIsoAPIKey                     string         `yaml:"AUTO_ISO_API_KEY"`
	DefinitionsCatalogURL             string         `yaml:"DEFINITIONS_CATALOG_URL"`
	DefinitionsWorkerURL              string         `yaml:"DEFINITIONS_WORKER_URL"`
	DefinitionsWorkerToken            string         `yaml:"DEFINITIONS_WORKER_TOKEN"`
	DatGroupURL                       url.URL        `yaml:"DAT_GROUP_URL"`
	DatGroupCustomerLogin             string         `yaml:"DAT_GROUP_CUSTOMER_LOGIN"`
	DatGroupCustomerNumber            string         `yaml:"DAT_GROUP_CUSTOMER_NUMBER"`
	DatGroupInterfacePartnerSignature string         `yaml:"DAT_GROUP_INTERFACE_PARTNER_SIGNATURE"`
	DatGroupCustomerSignature         string         `yaml:"DAT_GROUP_CUSTOMER_SIGNATURE"`
	SearchServiceAPIURL               url.URL        `yaml:"SEARCH_SERVICE_API_URL"`
	SearchServiceAPIKey               string         `yaml:"SEARCH_SERVICE_API_KEY"`
	SearchServiceIndexName            string         `yaml:"SEARCH_SERVICE_DEVICE_DEFINITION_INDEX"`
	JwtKeySetURL                      string         `yaml:"JWT_KEY_SET_URL"`
	IdentityAPIURL                    url.URL        `yaml:"IDENTITY_API_URL"`
	GoogleSheetsCredentials           string         `yaml:"GOOGLE_SHEETS_CREDENTIALS"`
	Japan17VINUser                    string         `yaml:"JAPAN17_VIN_USER"`
	Japan17VINPassword                string         `yaml:"JAPAN17_VIN_PASSWORD"`
	CarVxUserID                       string         `yaml:"CAR_VX_USER_ID"`
	CarVxAPIKey                       string         `yaml:"CAR_VX_API_KEY"`
	// used for Kaufmann API
	ElevaUsername string `yaml:"ELEVA_USERNAME"`
	ElevaPassword string `yaml:"ELEVA_PASSWORD"`
}

func (s *Settings) IsProd() bool {
	return s.Environment == "prod"
}
