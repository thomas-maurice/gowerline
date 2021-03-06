//nolint:unused
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Finnhub-Stock-API/finnhub-go"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/utils/cache"
	"go.uber.org/zap"
)

var (
	cfg            Config
	pluginName     = "finnhub"
	cachedData     map[string]finnhub.Quote
	stopChannel    chan bool
	stoppedChannel chan bool
	pluginConfig   *plugins.PluginConfig

	boltCache *cache.SimpleCache
)

const (
	DirectionUp     = "⬆️ "
	DirectionDown   = "⬇️ "
	cacheBucketName = "tickers"
)

type cachedTickerData struct {
	Timestamp time.Time      `json:"timestamp"`
	Quote     *finnhub.Quote `json:"quote"`
}

type Config struct {
	Token   string        `yaml:"token"`
	Tickers []string      `yaml:"tickers"`
	Refresh time.Duration `yaml:"refresh"`
}

type pluginArgs struct {
	Ticker           string `json:"ticker"`
	IncludeDirection bool   `json:"includeDirection"`
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
			var cached cachedTickerData
			found, err := boltCache.Get(ticker, &cached)
			if err != nil {
				log.Error("failed to fetch cached quote for ticker", zap.Error(err), zap.String("ticker", ticker))
				continue
			}
			log.Info("fetched data from cache", zap.String("ticker", ticker))
			if found {
				cachedData[ticker] = *cached.Quote
			}
			continue
		}

		cachedData[ticker] = quote
		err = boltCache.Put(ticker, &cachedTickerData{
			Timestamp: time.Now(),
			Quote:     &quote,
		})
		if err != nil {
			log.Error("could not cache result for ticker", zap.String("ticker", ticker), zap.Error(err))
		}
	}

	return nil
}

func run(log *zap.Logger) {
	err := updateTickers(log)
	if err != nil {
		log.Error("failed to update tickers", zap.Error(err))
	}

	tck := time.NewTicker(cfg.Refresh)

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

	err := pluginConfig.Config.Decode(&cfg)
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	for _, ticker := range cfg.Tickers {
		log.Info("added ticker", zap.String("ticker", ticker))
	}

	if cfg.Refresh < time.Second*60 {
		cfg.Refresh = time.Second * 60
	}

	log.Info(fmt.Sprintf("refreshing data every %v", cfg.Refresh))

	go run(log)

	return &types.PluginStartData{
		Metadata: types.PluginMetadata{
			Description: "Returns information about the stock price of certain tickers",
			Author:      "Thomas Maurice <thomas@maurice.fr>",
			Version:     "0.0.1",
			Functions: []types.FunctionDescriptor{
				{
					Name:        "ticker",
					Description: "Returns the stock price of a given ticket",
					Parameters: map[string]string{
						"ticker": "Symbol of the ticker to return",
					},
				},
			},
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

	content := fmt.Sprintf("%s: $%.02f", args.Ticker, quote.C)

	hlGroup := "gwl:ticker_generic"
	if quote.C > quote.Pc {
		hlGroup = "gwl:ticker_up"
		if args.IncludeDirection {
			content = fmt.Sprintf("%s %s", DirectionUp, content)
		}
	} else if quote.C < quote.Pc {
		hlGroup = "gwl:ticker_down"
		if args.IncludeDirection {
			content = fmt.Sprintf("%s %s", DirectionDown, content)
		}
	}

	return []*types.PowerlineReturn{
		{
			Content: content,
			HighlightGroup: []string{
				hlGroup,
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

	var err error

	boltCache, err = cache.NewSimpleCache(cacheBucketName, pCfg.BoltDB)

	//err := initCacheDB(pCfg.BoltDB)

	return &plugins.Plugin{
		Start: Start,
		Stop:  Stop,
		Call:  Call,
		Name:  pluginName,
	}, err
}

// noop main function
func main() {}
