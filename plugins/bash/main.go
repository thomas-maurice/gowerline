package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"sync"
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
	cachedData     map[string]string
	cacheMutex     *sync.Mutex
	runners        []*commandRunner
)

// define here the config of your plugin, if needed
// please use a file like `~/.gowerline/<pluginName>.yaml`
type Config struct {
	Commands map[string]struct {
		Cmd            string `yaml:"cmd"`
		Interval       int64  `yaml:"interval"`
		HighlightGroup string `yaml:"highlightGroup"`
	} `yaml:"commands"`
}

type commandRunner struct {
	StopChannel    chan bool
	StoppedChannel chan bool
	Interval       int64
	Cmd            string
	Name           string
}

func (c *commandRunner) run(log *zap.Logger) {
	tck := time.NewTicker(time.Duration(c.Interval) * time.Second)
	log.Info("registered command runner", zap.String("interval", (time.Duration(c.Interval)*time.Second).String()))
	for {
		select {
		case <-tck.C:
			err := c.runCommand(log)
			if err != nil {
				log.Error("could not run command", zap.Error(err))
			}
		case <-c.StopChannel:
			tck.Stop()
			log.Info("stopping command runner")
			c.StoppedChannel <- true
			return
		}
	}
}

// runCommand runs a command and caches the result
func (c *commandRunner) runCommand(log *zap.Logger) error {
	log.Debug("running command", zap.String("command", c.Cmd))
	cmd := exec.Command("/bin/bash", "-c", c.Cmd)
	stdout, err := cmd.Output()

	if err != nil {
		return err
	}

	c.cacheResult(strings.ReplaceAll(string(stdout), "\n", ""))
	return nil
}

func (c *commandRunner) cacheResult(result string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cachedData[c.Name] = result
}

// This is where you would get the plugin arguments passed
// to your plugin, this is what is contained in the `args`
// dictionary in the powerline config
type pluginArgs struct {
	CmdResult string `json:"cmd"`
}

// This is a loop you can populate if your plugin needs to do periodic data updates
// such as performing network calls or something.
func run(log *zap.Logger) {
	log.Info("starting main loop")

	for name, command := range cfg.Commands {
		runner := &commandRunner{
			Interval:       command.Interval,
			Name:           name,
			Cmd:            command.Cmd,
			StopChannel:    make(chan bool, 1),
			StoppedChannel: make(chan bool, 1),
		}

		runners = append(runners, runner)
	}

	log.Info("starting runners")

	wg := &sync.WaitGroup{}
	for _, r := range runners {
		wg.Add(1)
		runner := r
		go func() {
			runner.run(log.With(zap.String("command_name", runner.Name)))
			wg.Done()
		}()
	}

	<-stopChannel

	log.Info("terminating runners")

	for _, runner := range runners {
		log.Info("sending stop signal to runner", zap.String("runner", runner.Name))
		runner.StopChannel <- true
	}

	wg.Wait()
	stoppedChannel <- true
}

// Starts the plugin, here you might want to do all the initialisation you need
// load up config/tokens and what not, as well to start long running goroutines
// if your plugin requires it
func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	stopChannel = make(chan bool, 1)
	stoppedChannel = make(chan bool, 1)
	cachedData = make(map[string]string)
	cacheMutex = &sync.Mutex{}
	runners = make([]*commandRunner, 0)

	configBytes, err := ioutil.ReadFile(path.Join(pluginConfig.GowerlineDir, "bash.yaml"))
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
			"bash",
		},
	}, nil
}

// Stops anything you have started that is long runinng, like goroutines and what not
func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopping plugin",
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

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	data, ok := cachedData[args.CmdResult]
	if !ok {
		return nil, nil
	}

	hlgs := make([]string, 0)
	cmdCfg, ok := cfg.Commands[args.CmdResult]
	if ok {
		hlgs = append(hlgs, cmdCfg.HighlightGroup)
	}

	hlgs = append(hlgs, "information:regular")

	return []*types.PowerlineReturn{
		{
			Content:        data,
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
