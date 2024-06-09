package config

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	StaticPath string `json:"static_path"`
	AppPort    string `json:"port_app"`
}

func ReadConfig(filename string) (AppConfig, error) {
	var config AppConfig

	// Open the configuration file
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	// Create a new JSON decoder and decode the configuration
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}

	// Return the decoded configuration
	return config, nil
}
