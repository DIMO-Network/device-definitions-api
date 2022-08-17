package config

type Settings struct {
	Environment          string `yaml:"ENVIRONMENT"`
	Port                 string `yaml:"PORT"`
	LogLevel             string `yaml:"LOG_LEVEL"`
	DBUser               string `yaml:"DB_USER"`
	DBPassword           string `yaml:"DB_PASSWORD"`
	DBPort               string `yaml:"DB_PORT"`
	DBHost               string `yaml:"DB_HOST"`
	DBName               string `yaml:"DB_NAME"`
	DBMaxOpenConnections int    `yaml:"DB_MAX_OPEN_CONNECTIONS"`
	DBMaxIdleConnections int    `yaml:"DB_MAX_IDLE_CONNECTIONS"`
	ServiceName          string `yaml:"SERVICE_NAME"`
	ServiceVersion       string `yaml:"SERVICE_VERSION"`
	GRPC_Port            string `yaml:"GRPC_PORT"`
	TraceMonitorView     string `yaml:"TRACE_MONITOR_VIEW"`
}
