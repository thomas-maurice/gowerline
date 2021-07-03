package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port    int64    `yaml:"port"`
	Plugins []string `yaml:"plugins"`
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