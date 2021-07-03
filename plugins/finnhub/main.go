package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/Finnhub-Stock-API/finnhub-go"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	cfg            Config
	pluginName     = "finnhub"
	cachedData     map[string]finnhub.Quote
	stopChannel    chan bool
	stoppedChannel chan bool
	pluginConfig   *plugins.PluginConfig
)

type Config struct {
	Token   string   `yaml:"token"`
	Tickers []string `yaml:"tickers"`
}

type pluginArgs struct {
	Ticker string `json:"ticker"`
}

// updatesTickers gets the data for caching
func updateTickers(log *zap.Logger) error {
	log.Info("updating ticker data")
	client := finnhub.NewAPIClient(finnhub.NewConfiguration()).DefaultApi
	ctx := context.WithValue(context.Background(), finnhub.ContextAPIKey, finnhub.APIKey{
		Key: cfg.Token,
	})

	for _, ticker := range cfg.Tickers {
		quote, _, err := client.Quote(ctx, ticker)
		if err != nil {
			log.Error("failed to fetch quote for ticker", zap.Error(err), zap.String("ticker", ticker))
			continue
		}

		cachedData[ticker] = quote
	}

	return nil
}

func run(log *zap.Logger) {
	err := updateTickers(log)
	if err != nil {
		log.Error("failed to update tickers", zap.Error(err))
	}

	tck := time.NewTicker(time.Minute)

	for {
		select {
		case <-tck.C:
			err := updateTickers(log)
			if err != nil {
				log.Error("failed to update tickers", zap.Error(err))
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
	cachedData = make(map[string]finnhub.Quote)
	stopChannel = make(chan bool)
	stoppedChannel = make(chan bool)

	configBytes, err := ioutil.ReadFile(path.Join(pluginConfig.GowerlineDir, "finnhub.yaml"))
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	err = yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	for _, ticker := range cfg.Tickers {
		log.Info("added ticker", zap.String("ticker", ticker))
	}

	go run(log)

	return &types.PluginStartData{
		Functions: []string{
			"ticker",
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
// check what is called using the payload.Function attribute
func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	var args pluginArgs
	err := json.Unmarshal(*payload.Args, &args)
	if err != nil {
		log.Error("could not unmarshal arguments", zap.Error(err))
		return nil, err
	}

	quote, ok := cachedData[args.Ticker]
	if !ok {
		return nil, nil
	}

	hlGroup := "gwl:ticker_generic"
	if quote.C > quote.Pc {
		hlGroup = "gwl:ticker_up"
	} else if quote.C < quote.Pc {
		hlGroup = "gwl:ticker_down"
	}

	return []*types.PowerlineReturn{
		{
			Content: fmt.Sprintf("%s: $%.02f", args.Ticker, quote.C),
			HighlightGroup: []string{
				hlGroup,
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
		Name:  pluginName,
	}, nil
}
