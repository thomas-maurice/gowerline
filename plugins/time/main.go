package main

import (
	"context"
	"fmt"
	"time"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
)

func init() {

}

var (
	pluginName = "time"
)

// Starts the plugin, here you might want to do all the initialisation you need
// load up config/tokens and what not, as well to start long running goroutines
// if your plugin requires it
func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	log.Info(
		"started plugin",
		zap.String("plugin", pluginName),
	)
	return &types.PluginStartData{
		Functions: []string{
			"time",
		},
	}, nil
}

// Stops anything you have started that is long runinng, like goroutines and what not
func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopped plugin",
		zap.String("plugin", pluginName),
	)
	return nil
}

// Returns the actual segment iself. If your plugin handles different functions you should
// check what is called using the payload.Function attribute
func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	log.Info(
		"called plugin",
		zap.String("plugin", pluginName),
	)
	return []*types.PowerlineReturn{
		{
			Content: fmt.Sprintf("%d", time.Now().Unix()),
			HighlightGroup: []string{
				"hostname", // using the hostname HL group because it is likely to be defined in the conf
			},
		},
	}, nil
}

// Init builds and returns the plugin itself
func Init(ctx context.Context, log *zap.Logger) (*plugins.Plugin, error) {
	log.Info(
		"loaded plugin",
		zap.String("plugin", pluginName),
	)
	return &plugins.Plugin{
		Start: Start,
		Stop:  Stop,
		Call:  Call,
		Name:  "time",
	}, nil
}
