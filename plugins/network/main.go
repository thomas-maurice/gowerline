//nolint:unused
package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
)

const (
	defaultPublicIpService = "https://checkip.amazonaws.com/"
)

var (
	cfg             Config
	stopChannel     chan bool
	stoppedChannel  chan bool
	pluginConfig    *plugins.PluginConfig
	publicIpAddress string
)

type Config struct {
	IpService string `json:"ipService" yaml:"ipService"`
}

type pluginArgs struct {
}

func update(log *zap.Logger) error {
	log.Info("running the update loop")

	if cfg.IpService == "" {
		cfg.IpService = defaultPublicIpService
	}

	resp, err := http.Get(cfg.IpService)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	publicIpAddress = strings.ReplaceAll(string(b), "\n", "")

	return nil
}

func run(log *zap.Logger) {
	err := update(log)
	if err != nil {
		log.Error("failed to run plugin data refresh", zap.Error(err))
	}

	tck := time.NewTicker(time.Minute)

	for {
		select {
		case <-tck.C:
			err = update(log)
			if err != nil {
				log.Error("could not update data", zap.Error(err))
			}
		case <-stopChannel:
			stoppedChannel <- true
			return
		}
	}
}

func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	stopChannel = make(chan bool)
	stoppedChannel = make(chan bool)

	err := pluginConfig.Config.Decode(&cfg)
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	go run(log)

	return &types.PluginStartData{
		Metadata: types.PluginMetadata{
			Description: "Gather information about your network connectivity",
			Author:      "Thomas Maurice <thomas@maurice.fr>",
			Version:     "devel",
			Functions: []types.FunctionDescriptor{
				{
					Name:        "public_ip",
					Description: "Returns your public IP address",
					Parameters:  map[string]string{},
				},
			},
		},
	}, nil
}

func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopped plugin",
	)

	stopChannel <- true
	<-stoppedChannel

	return nil
}

func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	var args pluginArgs
	err := json.Unmarshal(*payload.Args, &args)
	if err != nil {
		log.Error("could not unmarshal plugin arguments", zap.Error(err))
		return nil, err
	}

	switch payload.Function {
	case "public_ip":
		return []*types.PowerlineReturn{
			{
				Content: publicIpAddress,
				HighlightGroup: []string{
					"gwl:public_ip",
				},
			},
		}, nil
	default:
		return []*types.PowerlineReturn{
			{
				Content: "no such function",
				HighlightGroup: []string{
					"information:regular",
				},
			},
		}, nil
	}
}

func Init(ctx context.Context, log *zap.Logger, pCfg *plugins.PluginConfig) (*plugins.Plugin, error) { //nolint:deadcode
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
