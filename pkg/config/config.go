package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	System struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Env     string `yaml:"env"`
	} `yaml:"system"`
	CoreServices     []string `yaml:"core_services"`
	OptionalServices []string `yaml:"optional_services"`
	ExternalModules  map[string]struct {
		Mode     string `yaml:"mode"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"external_modules"`
	Discovery struct {
		RegistryURL       string `yaml:"registry_url"`
		HeartbeatInterval string `yaml:"heartbeat_interval"`
		Timeout           string `yaml:"timeout"`
	} `yaml:"discovery"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
