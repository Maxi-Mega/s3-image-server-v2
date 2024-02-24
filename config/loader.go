package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(configPath string) (Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, err //nolint:wrapcheck
	}

	defer file.Close()

	var cfg Config

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
