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
	rootCmd.PersistentFlags().StringVarP(&pluginsDir, "plugins", "p", defaultPluginDir, "Default plugin directory")
	rootCmd.PersistentFlags().StringVarP(&marshaller, "marshaller", "m", "yaml", "Marshaller for server responses (client mode)")

	initServerCmd()
	rootCmd.AddCommand(serverCmd)
}
