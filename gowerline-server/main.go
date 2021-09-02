package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"github.com/thomas-maurice/gowerline/gowerline-server/handlers"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	configFile string
	pluginDir  string
	homeDir    string
)

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultPluginDir := path.Join(homeDir, ".gowerline", "plugins")

	flag.StringVar(&pluginDir, "plugins", defaultPluginDir, "directory the plugins are at")
	flag.StringVar(&configFile, "config", path.Join(homeDir, ".gowerline", "server.yaml"), "config file")
}

func main() {
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	log.Info(
		"starting gowerline server",
		zap.String("version", version.Version),
		zap.String("build_host", version.BuildHost),
		zap.String("build_time", version.BuildTime),
		zap.String("build_hash", version.BuildHash),
	)

	cfg, err := config.NewConfigFromFile(configFile)
	if err != nil {
		log.Panic("could not load config", zap.Error(err))
	}

	// TODO: use this to control timeouts
	ctx := context.Background()

	pluginMap := make(map[string]*plugins.Plugin)
	pluginList := make([]*plugins.Plugin, 0)
	for _, plgName := range cfg.Plugins {
		plgPath := path.Join(pluginDir, plgName)
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

	ginLoggerConfig := zap.NewProductionConfig()
	ginLoggerConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	ginLogger, err := ginLoggerConfig.Build()
	if err != nil {
		log.Panic("could not setup gin logger", zap.Error(err))
	}
	r.Use(ginzap.Ginzap(ginLogger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(log, true))

	err = handlers.SetupHandlers(r, ctx, log, pluginMap)
	if err != nil {
		log.Panic("could not setup handlers", zap.Error(err))
	}

	go func() {
		if cfg.Listen.Unix != "" {
			var listenPath string
			currentUser, err := user.Current()
			if err != nil {
				log.Panic("could not get current user", zap.Error(err))
			}
			homeDir := currentUser.HomeDir

			if strings.HasPrefix(cfg.Listen.Unix, "~/") {
				listenPath = filepath.Join(homeDir, cfg.Listen.Unix[2:])
			}

			log.Info("listening on an unix socket", zap.String("socket", listenPath))
			os.Remove(listenPath)
			err = r.RunUnix(listenPath)
			if err != nil {
				log.Panic("could not listen", zap.Error(err))
			}

			log.Info("closed unix socket")

			return
		}

		err = r.Run(fmt.Sprintf("127.0.0.1:%d", cfg.Listen.Port))
		if err != nil {
			log.Panic("could not listen", zap.Error(err))
		}
		log.Info("closed socket")
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
