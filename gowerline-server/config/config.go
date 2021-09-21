package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen struct {
		Port int64  `yaml:"port"`
		Unix string `yaml:"unix"`
	} `yaml:"listen"`
	Plugins []string `yaml:"plugins"`
	Debug   bool     `yaml:"debug"`
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
