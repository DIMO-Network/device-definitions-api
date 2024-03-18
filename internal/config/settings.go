package config

import (
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/shared/redis"
)

type Settings struct {
	Environment                    string         `yaml:"ENVIRONMENT"`
	Port                           string         `yaml:"PORT"`
	LogLevel                       string         `yaml:"LOG_LEVEL"`
	DB                             db.Settings    `yaml:"DB"`
	ServiceName                    string         `yaml:"SERVICE_NAME"`
	ServiceVersion                 string         `yaml:"SERVICE_VERSION"`
	GRPCPort                       string         `yaml:"GRPC_PORT"`
	TraceMonitorView               string         `yaml:"TRACE_MONITOR_VIEW"`
	ElasticSearchAppSearchHost     string         `yaml:"ELASTIC_SEARCH_APP_SEARCH_HOST"`
	ElasticSearchAppSearchToken    string         `yaml:"ELASTIC_SEARCH_APP_SEARCH_TOKEN"`
	ElasticSearchDeviceStatusHost  string         `yaml:"ELASTIC_SEARCH_DEVICE_STATUS_HOST"`
	ElasticSearchDeviceStatusToken string         `yaml:"ELASTIC_SEARCH_DEVICE_STATUS_TOKEN"`
	IPFSNodeEndpoint               string         `yaml:"IPFS_NODE_ENDPOINT"`
	DrivlyAPIKey                   string         `yaml:"DRIVLY_API_KEY"`
	DrivlyVINAPIURL                string         `yaml:"DRIVLY_VIN_API_URL"`
	DrivlyOfferAPIURL              string         `yaml:"DRIVLY_OFFER_API_URL"`
	Redis                          redis.Settings `yaml:"REDIS"`
	FuelAPIVehiclesEndpoint        string         `yaml:"FUEL_API_SEARCH_VEHICLES_ENDPOINT"`
	FuelAPIImagesEndpoint          string         `yaml:"FUEL_API_SEARCH_IMAGES_ENDPOINT"`
	FuelAPIKey                     string         `yaml:"FUEL_API_KEY"`
	NHTSARecallsFileURL            string         `yaml:"NHTSA_RECALLS_FILE_URL"`
	VincarioAPIURL                 string         `yaml:"VINCARIO_API_URL"`
	VincarioAPIKey                 string         `yaml:"VINCARIO_API_KEY"`
	VincarioAPISecret              string         `yaml:"VINCARIO_API_SECRET"`
	AutoIsoAPIUid                  string         `yaml:"AUTO_ISO_API_UID"`
	AutoIsoAPIKey                  string         `yaml:"AUTO_ISO_API_KEY"`
	EthereumRPCURL                 string         `yaml:"ETHEREUM_RPC_URL"`
	SenderPrivateKey               string         `yaml:"SENDER_PRIVATE_KEY"`
	PrivateKeyMode                 bool           `yaml:"PRIVATE_KEY_MODE"`
	KMSKeyID                       string         `yaml:"KMS_KEY_ID"`
	EthereumRegistryAddress        string         `yaml:"ETHEREUM_REGISTRY_ADDRESS"`
	EthereumSendTransaction        bool           `yaml:"ETHEREUM_SEND_TRANSACTION"`
}

func (s *Settings) IsProd() bool {
	return s.Environment == "prod"
}
