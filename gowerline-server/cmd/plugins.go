package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/utils"
	"go.uber.org/zap"
)

var (
	runArgs []string
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
	Use:   "functions [plugin]",
	Short: "Retrieves info about a specific plugin's functions",
	Args:  cobra.ExactArgs(1),
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

var pluginFunctionRunCmd = &cobra.Command{
	Use:   "function-run [function]",
	Short: "Runs a function with the given parameters",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewConfigFromFile(configFile)
		if err != nil {
			log.Panic("could not load config", zap.Error(err))
		}

		client := utils.NewHTTPClientFromConfig(cfg)

		argsMap := make(map[string]string)
		for _, arg := range runArgs {
			splitted := strings.Split(arg, "=")
			if len(splitted) == 0 {
				continue
			} else if len(splitted) == 1 {
				argsMap[splitted[0]] = ""
			} else if len(splitted) == 2 {
				argsMap[splitted[0]] = splitted[1]
			} else {
				argsMap[splitted[0]] = strings.Join(splitted[1:], "=")
			}
		}

		var payload types.Payload
		payload.Function = args[0]

		b, err := json.Marshal(argsMap)
		if err != nil {
			log.Fatal("could not marshal args", zap.Error(err))
		}

		msg := json.RawMessage(b)
		payload.Args = &msg
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("could not gte current working directory", zap.Error(err))
		}

		payload.Cwd = cwd
		env := os.Environ()
		payload.Env = make(map[string]string)
		for _, envVar := range env {
			splitted := strings.Split(envVar, "=")
			if len(splitted) == 0 {
				continue
			} else if len(splitted) == 1 {
				payload.Env[splitted[0]] = ""
			} else if len(splitted) == 2 {
				payload.Env[splitted[0]] = splitted[1]
			} else {
				payload.Env[splitted[0]] = strings.Join(splitted[1:], "=")
			}
		}

		b, err = json.Marshal(payload)
		if err != nil {
			log.Fatal("could not marshal request", zap.Error(err))
		}

		if debug {
			output(payload)
		}

		resp, err := client.Post(utils.BaseURLFromConfig(cfg)+"/plugin", "application/json", strings.NewReader(string(b)))
		if err != nil {
			log.Fatal("could not fetch the server's version", zap.Error(err))
		}
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("could not read http response", zap.Error(err))
		}
		defer resp.Body.Close()

		content := make([]types.PowerlineReturn, 0)
		err = json.Unmarshal(b, &content)
		if err != nil {
			log.Fatal("could not unmarshal server response", zap.Error(err))
		}

		output(content)
	},
}

func initPluginCommand() {
	pluginFunctionRunCmd.PersistentFlags().StringSliceVarP(&runArgs, "arg", "a", []string{}, "Arguments to pass in a key=value format")

	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginFunctionsCmd)
	pluginCmd.AddCommand(pluginFunctionRunCmd)
}
