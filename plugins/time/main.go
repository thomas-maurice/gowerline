//nolint:unused
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
)

var (
	pluginConfig *plugins.PluginConfig
)

// Starts the plugin, here you might want to do all the initialisation you need
// load up config/tokens and what not, as well to start long running goroutines
// if your plugin requires it
func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	log.Info(
		"started plugin",
	)

	/*
		// If it even has a config
		if pluginConfig.Config != nil {
			err := pluginConfig.Config.Decode(&cfg)
			if err != nil {
				log.Panic("could not load configuration", zap.Error(err))
			}
		}
	*/

	return &types.PluginStartData{
		Metadata: types.PluginMetadata{
			Description: "Shows time, it is a debug segment",
			Author:      "Thomas Maurice <thomas@maurice.fr>",
			Version:     "0.0.1",
			Functions: []types.FunctionDescriptor{
				{Name: "time", Description: "Displays the current tine"},
			},
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
	return []*types.PowerlineReturn{
		{
			Content: fmt.Sprintf("%d", time.Now().Unix()),
			HighlightGroup: []string{
				"information:regular",
			},
		},
	}, nil
}

// Init builds and returns the plugin itself
func Init(ctx context.Context, log *zap.Logger, pCfg *plugins.PluginConfig) (*plugins.Plugin, error) { //nolint:deadcode
	log.Info(
		"loaded plugin",
	)

	pluginConfig = pCfg

	return &plugins.Plugin{
		Start: Start,
		Stop:  Stop,
		Call:  Call,
		Name:  pluginConfig.PluginName,
	}, nil
}

// noop main function
func main() {}
