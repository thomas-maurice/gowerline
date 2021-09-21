package cmd

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func output(obj interface{}) {
	switch marshaller {
	case "json":
		b, err := json.MarshalIndent(&obj, "", "  ")
		if err != nil {
			log.Fatal("could not marshall struct", zap.Error(err))
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(&obj)
		if err != nil {
			log.Fatal("could not marshall struct", zap.Error(err))
		}
		fmt.Println(string(b))
	default:
		fmt.Println(obj)
	}
}
