package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type ConfigPlugin struct {
	Name     string    `yaml:"name"`
	Disabled bool      `yaml:"disabled"`
	Config   yaml.Node `yaml:"config"`
}

type Config struct {
	Listen struct {
		Port int64  `yaml:"port"`
		Unix string `yaml:"unix"`
	} `yaml:"listen"`
	Debug   bool           `yaml:"debug"`
	Plugins []ConfigPlugin `yaml:"plugins"`
}

func NewConfigFromFile(configFile string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(configBytes, &cfg)
	return &cfg, err
}
