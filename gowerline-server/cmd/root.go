package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFile string
	pluginsDir string
	homeDir    string
	marshaller string
	log        *zap.Logger
	debug      bool
)

var rootCmd = &cobra.Command{
	Use:   "gowerline",
	Short: "generate powerline segments from Go !",
	Long:  ``,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultPluginDir := path.Join(homeDir, ".gowerline", "plugins")
	defaultConfigFile := path.Join(homeDir, ".gowerline", "gowerline.yaml")

	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigFile, "Default config file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Toggle debug mode")
	rootCmd.PersistentFlags().StringVarP(&pluginsDir, "plugins", "p", defaultPluginDir, "Default plugin directory")
	rootCmd.PersistentFlags().StringVarP(&marshaller, "output", "o", "json", "Output format for server responses (client mode), must be yaml or json")

	initServerCmd()
	initPluginCommand()

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pluginCmd)
}
