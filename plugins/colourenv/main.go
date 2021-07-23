package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"
	"regexp"

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

type ColourConfig struct {
	Regex          string         `yaml:"regex"`
	HighlightGroup string         `yaml:"highlightGroup"`
	CompiledRegex  *regexp.Regexp `yaml:"-"`
}

type Config struct {
	Variables map[string][]ColourConfig `yaml:"variables"`
}

func (c *Config) Compile(log *zap.Logger) error {
	for variable, regexes := range c.Variables {
		for idx, regex := range regexes {
			compiledRegex, err := regexp.Compile(regex.Regex)
			if err != nil {
				log.Error("could not compile regex", zap.String("variable", variable), zap.String("regex", regex.Regex))
				return err
			}
			log.Info("adding regex for variable", zap.String("variable", variable), zap.String("regex", regex.Regex))
			c.Variables[variable][idx].CompiledRegex = compiledRegex
		}
	}
	return nil
}

func (c *Config) GetHighlights(log *zap.Logger, variable string, value string) []string {
	colourConfig, ok := c.Variables[variable]
	if !ok {
		return nil
	}

	for _, cfg := range colourConfig {

		if cfg.CompiledRegex.Match([]byte(value)) {
			return []string{cfg.HighlightGroup}
		}
	}

	return nil
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

	err = cfg.Compile(log)
	if err != nil {
		log.Panic("failed to compile regexes", zap.Error(err))
	}

	for k, v := range cfg.Variables {
		for _, cf := range v {
			log.Info("added variable", zap.String("variable", k), zap.String("regex", cf.Regex), zap.String("highlight_group", cf.HighlightGroup))
		}
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

	hlgs := cfg.GetHighlights(log, args.Variable, val)
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
