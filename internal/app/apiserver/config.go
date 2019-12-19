package apiserver

// Config ...
type Config struct {
	BindAddr    string `toml:"bind_addr"`
	DatabaseURL string `toml:"database_url"`
}

// NewConfig - create config with default values
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
	}
}
