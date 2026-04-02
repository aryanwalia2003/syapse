package config

type ServerConfig struct {
	Port int    `env:"PORT" envDefault:"8080"`
	Env  string `env:"ENV" envDefault:"development"`
}
type DatabaseConfig struct {
	PocketBaseURL string `env:"PB_URL,required"`
	PostgresURL   string `env:"PG_URL"`
}

type SynapseConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
}
