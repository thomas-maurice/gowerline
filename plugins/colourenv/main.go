package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	cfg          Config
	pluginConfig *plugins.PluginConfig
	defaultHLG   = "information:regular"
)

type Config struct {
	Variables map[string]map[string]string `yaml:"variables"`
}

type pluginArgs struct {
	Variable string `json:"variable"`
}

func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	configBytes, err := ioutil.ReadFile(path.Join(pluginConfig.GowerlineDir, "colourenv.yaml"))
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	err = yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	for k, v := range cfg.Variables {
		log.Info("added variable", zap.String("variable", k), zap.Any("highlights_groups", v))
	}

	return &types.PluginStartData{
		Functions: []string{
			"colourenv",
		},
	}, nil
}

// Stops anything you have started that is long runinng, like goroutines and what not
func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopped plugin",
	)

	return nil
}

// Returns the actual segment iself. If your plugin handles different functions you should
// check what is called using the payload.Function attribute
func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	var args pluginArgs
	err := json.Unmarshal(*payload.Args, &args)
	if err != nil {
		log.Error("could not unmarshal arguments", zap.Error(err))
		return nil, err
	}

	val, ok := payload.Env[args.Variable]
	if !ok {
		return nil, nil
	}

	hlgs := make([]string, 0)
	varMapping, ok := cfg.Variables[args.Variable]
	if !ok {
		return nil, nil
	}

	hlg, ok := varMapping[val]
	if ok {
		hlgs = append(hlgs, hlg)
	}

	hlgs = append(hlgs, defaultHLG)

	return []*types.PowerlineReturn{
		{
			Content:        val,
			HighlightGroup: hlgs,
		},
	}, nil
}

// Init builds and returns the plugin itself
func Init(ctx context.Context, log *zap.Logger, pCfg *plugins.PluginConfig) (*plugins.Plugin, error) {
	log.Info(
		"loaded plugin",
	)

	pluginConfig = pCfg

	return &plugins.Plugin{
		Start: Start,
		Stop:  Stop,
		Call:  Call,
		Name:  pCfg.PluginName,
	}, nil
}
