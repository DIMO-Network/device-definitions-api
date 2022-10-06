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
}
