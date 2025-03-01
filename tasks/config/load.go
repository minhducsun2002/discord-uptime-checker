package config

import (
	"discord-uptime-checker/constants"
	"discord-uptime-checker/structures"
	"gopkg.in/yaml.v3"
	"os"
)

func LoadConfig() (structures.Config, error) {
	file, err := os.ReadFile(constants.ConfigFileName)
	if err != nil {
		return nil, err
	}

	var config structures.Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return config, nil
}
