package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ModelsDir  string   `mapstructure:"models_dir"`
	InputCol   string   `mapstructure:"input_col"`
	Categories []string `mapstructure:"categories"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
