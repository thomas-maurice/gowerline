package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/utils"
	"go.uber.org/zap"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Interact with plugins",
	Long:  ``,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists plugins currently loaded",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewConfigFromFile(configFile)
		if err != nil {
			log.Panic("could not load config", zap.Error(err))
		}

		client := utils.NewHTTPClientFromConfig(cfg)

		resp, err := client.Get(utils.BaseURLFromConfig(cfg) + "/plugins")
		if err != nil {
			log.Fatal("could not fetch the server's version", zap.Error(err))
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("could not read http response", zap.Error(err))
		}
		defer resp.Body.Close()

		pluginInfo := make(map[string]types.PluginMetadata)
		err = json.Unmarshal(b, &pluginInfo)
		if err != nil {
			log.Fatal("could not unmarshal server response", zap.Error(err))
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Description", "Version", "Author"})

		for name, meta := range pluginInfo {
			table.Append([]string{name, meta.Description, meta.Author, meta.Version})
		}
		table.Render()
	},
}

var pluginFunctionsCmd = &cobra.Command{
	Use:   "functions",
	Short: "Retrieves info about a specific plugin's functions",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("you must pass the plugin's name")
		}

		cfg, err := config.NewConfigFromFile(configFile)
		if err != nil {
			log.Panic("could not load config", zap.Error(err))
		}

		client := utils.NewHTTPClientFromConfig(cfg)

		resp, err := client.Get(utils.BaseURLFromConfig(cfg) + "/plugins")
		if err != nil {
			log.Fatal("could not fetch the server's version", zap.Error(err))
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("could not read http response", zap.Error(err))
		}
		defer resp.Body.Close()

		pluginInfo := make(map[string]types.PluginMetadata)
		err = json.Unmarshal(b, &pluginInfo)
		if err != nil {
			log.Fatal("could not unmarshal server response", zap.Error(err))
		}

		for name, meta := range pluginInfo {
			if name == args[0] {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Function name", "Description", "Argument", "Argument help"})
				for _, functionMeta := range meta.Functions {
					table.Append([]string{functionMeta.Name, functionMeta.Description, "", ""})
					for paramName, paramDesc := range functionMeta.Parameters {
						table.Append([]string{"", "", paramName, paramDesc})
					}
				}
				table.Render()
				return
			}
		}

		log.Fatal("no such plugin", zap.String("plugin", args[0]))
	},
}

func initPluginCommand() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginFunctionsCmd)
}
