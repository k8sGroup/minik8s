package config

type Config struct {
}

type CompletedConfig struct {
	*Config
}

func (c *Config) Complete() *CompletedConfig {
	// TODO : complete config
	return &CompletedConfig{c}
}
