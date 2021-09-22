package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	"github.com/spf13/cobra"
	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"github.com/thomas-maurice/gowerline/gowerline-server/handlers"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/utils"
	"github.com/thomas-maurice/gowerline/gowerline-server/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manages the server",
	Long:  ``,
}

var serverRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
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
			zap.String("target_os", version.OS),
			zap.String("target_arch", version.Arch),
		)

		cfg, err := config.NewConfigFromFile(configFile)
		if err != nil {
			log.Panic("could not load config", zap.Error(err))
		}

		// TODO: use this to control timeouts
		ctx := context.Background()

		pluginMap := make(map[string]*plugins.Plugin)
		pluginList := make([]*plugins.Plugin, 0)
		for _, plgCfg := range cfg.Plugins {
			if plgCfg.Disabled {
				log.Info("skipping disabled pluigin", zap.String("plugin", plgCfg.Name))
				continue
			}

			plgPath := path.Join(pluginsDir, plgCfg.Name)
			plg, err := plugins.NewPlugin(ctx, log, plgPath, &plugins.PluginConfig{
				UserHome:     homeDir,
				GowerlineDir: path.Join(homeDir, ".gowerline"),
				PluginName:   plgCfg.Name,
				Config:       plgCfg.Config,
			})
			if err != nil {
				log.Panic(fmt.Sprintf("could not load plugin %s", plgCfg.Name), zap.Error(err))
			}
			startData, err := plg.RunStart(ctx, log)
			if err != nil {
				log.Panic(fmt.Sprintf("could not load plugin %s", plgCfg.Name), zap.Error(err))
			}

			plg.Metadata = startData.Metadata

			log.Info(
				"loaded plugin",
				zap.String("plugin", plgCfg.Name),
				zap.String("version", startData.Metadata.Version),
				zap.String("author", startData.Metadata.Author),
			)

			for _, fn := range startData.Metadata.Functions {
				log.Info(
					"registered new function for plugin",
					zap.String("plugin", plgCfg.Name),
					zap.String("function", fn.Name),
				)
				pluginMap[fn.Name] = plg
			}
			pluginList = append(pluginList, plg)
		}

		r := gin.New()

		ginLoggerConfig := zap.NewProductionConfig()
		if cfg.Debug {
			ginLoggerConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		} else {
			ginLoggerConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		}
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
	},
}

var serverVersionCmd = &cobra.Command{
	Use:   "server-version",
	Short: "Returns the version of the server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewConfigFromFile(configFile)
		if err != nil {
			log.Panic("could not load config", zap.Error(err))
		}

		client := utils.NewHTTPClientFromConfig(cfg)

		resp, err := client.Get(utils.BaseURLFromConfig(cfg) + "/version")
		if err != nil {
			log.Fatal("could not fetch the server's version", zap.Error(err))
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("could not read http response", zap.Error(err))
		}
		defer resp.Body.Close()

		var serverInfo types.ServerVersionInfo
		err = json.Unmarshal(b, &serverInfo)
		if err != nil {
			log.Fatal("could not unmarshal server response", zap.Error(err))
		}

		output(serverInfo)
	},
}

func initServerCmd() {
	serverCmd.AddCommand(serverRunCmd)
}
