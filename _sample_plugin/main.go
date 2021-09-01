package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"
	"time"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	cfg            Config
	stopChannel    chan bool
	stoppedChannel chan bool
	pluginConfig   *plugins.PluginConfig
)

// define here the config of your plugin, if needed
// please use a file like `~/.gowerline/<pluginName>.yaml`
type Config struct {
	SomeVariable string `json:"someVariable"`
}

// This is where you would get the plugin arguments passed
// to your plugin, this is what is contained in the `args`
// dictionary in the powerline config
type pluginArgs struct {
	SomeVariable string `json:"someVariable"`
}

// update gets the data for caching
func update(log *zap.Logger) error {
	log.Info("running the update loop")

	return nil
}

// This is a loop you can populate if your plugin needs to do periodic data updates
// such as performing network calls or something.
func run(log *zap.Logger) {
	err := update(log)
	if err != nil {
		log.Error("failed to run plugin data refresh", zap.Error(err))
	}

	tck := time.NewTicker(time.Minute)

	for {
		select {
		case <-tck.C:
			// Do something here
			update(log)
			if err != nil {
				log.Error("could not update data", zap.Error(err))
			}
		case <-stopChannel:
			stoppedChannel <- true
			return
		}
	}
}

// Starts the plugin, here you might want to do all the initialisation you need
// load up config/tokens and what not, as well to start long running goroutines
// if your plugin requires it
func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	stopChannel = make(chan bool)
	stoppedChannel = make(chan bool)

	configBytes, err := ioutil.ReadFile(path.Join(pluginConfig.GowerlineDir, "your_plugin.yaml"))
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	err = yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}
	go run(log)

	return &types.PluginStartData{
		Functions: []string{
			"some_function",
		},
	}, nil
}

// Stops anything you have started that is long runinng, like goroutines and what not
func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopped plugin",
	)

	stopChannel <- true
	<-stoppedChannel

	return nil
}

// Returns the actual segment iself. If your plugin handles different functions you should
// check which one is called using the `payload.Function` attribute
func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	var args pluginArgs
	err := json.Unmarshal(*payload.Args, &args)
	if err != nil {
		log.Error("could not unmarshal plugin arguments", zap.Error(err))
		return nil, err
	}

	return []*types.PowerlineReturn{
		{
			Content: "bonjour",
			HighlightGroup: []string{
				// You add personalised highlight groups
				// but you should put them in the README
				// so users can adapt their themes.
				//
				// as a safety measure you can add a default
				// HLG like this even though the extension
				// should already do it.
				"information:regular",
			},
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
