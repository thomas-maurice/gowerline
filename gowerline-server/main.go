package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"github.com/thomas-maurice/gowerline/gowerline-server/handlers"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"go.uber.org/zap"
)

var (
	configFile string
	homeDir    string
)

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&configFile, "config", path.Join(homeDir, ".gowerline", "server.yaml"), "config file")
}

func main() {
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	cfg, err := config.NewConfigFromFile(configFile)
	if err != nil {
		log.Panic("could not load config", zap.Error(err))
	}

	// TODO: use this to control timeouts
	ctx := context.Background()

	pluginMap := make(map[string]*plugins.Plugin)
	pluginList := make([]*plugins.Plugin, 0)
	for _, plgName := range cfg.Plugins {
		plgPath := path.Join(homeDir, ".gowerline", "plugins", plgName)
		plg, err := plugins.NewPlugin(ctx, log, plgPath, &plugins.PluginConfig{
			UserHome:     homeDir,
			GowerlineDir: path.Join(homeDir, ".gowerline"),
			PluginName:   plgName,
		})
		if err != nil {
			log.Panic(fmt.Sprintf("could not load plugin %s", plgName), zap.Error(err))
		}
		data, err := plg.RunStart(ctx, log)
		if err != nil {
			log.Panic(fmt.Sprintf("could not load plugin %s", plgName), zap.Error(err))
		}
		for _, fn := range data.Functions {
			log.Info(
				"registered new function for plugin",
				zap.String("plugin", plgName),
				zap.String("function", fn),
			)
			pluginMap[fn] = plg
		}
		pluginList = append(pluginList, plg)
	}

	r := gin.New()

	r.Use(ginzap.Ginzap(log, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(log, true))

	err = handlers.SetupHandlers(r, ctx, log, pluginMap)
	if err != nil {
		log.Panic("could not setup handler", zap.Error(err))
	}

	go func() {
		err = r.Run(fmt.Sprintf("127.0.0.1:%d", cfg.Port))
		if err != nil {
			log.Panic("could not listen", zap.Error(err))
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	sig := <-signalChan
	log.Info("caught signal", zap.String("signal, exiting", sig.String()))
	for _, plg := range pluginList {
		log.Info("stopping plugin", zap.String("plugin", plg.Name))
		err := plg.RunStop(ctx, log)
		if err != nil {
			log.Error("failed to stop plugin", zap.String("plugin", plg.Name))
		}
	}
}
