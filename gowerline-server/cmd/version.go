package cmd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/utils"
	"github.com/thomas-maurice/gowerline/gowerline-server/version"
	"go.uber.org/zap"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns the version of the binary",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		data := make(map[string]interface{})

		data["client_version"] = types.ServerVersionInfo{
			BuildHost:       version.BuildHost,
			BuildTime:       version.BuildTime,
			GitHash:         version.BuildHash,
			Version:         version.Version,
			OperatingSystem: version.OS,
			Architecture:    version.Arch,
		}

		cfg, err := config.NewConfigFromFile(configFile)
		if err != nil {
			data["server_version"] = err.Error()
			output(data)
			log.Panic("could not load config", zap.Error(err))
		}

		client := utils.NewHTTPClientFromConfig(cfg)

		resp, err := client.Get(utils.BaseURLFromConfig(cfg) + "/version")
		if err != nil {
			data["server_version"] = err.Error()
			output(data)
			log.Fatal("could not fetch the server's version", zap.Error(err))
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			data["server_version"] = err.Error()
			output(data)
			log.Fatal("could not read http response", zap.Error(err))
		}
		defer resp.Body.Close()

		var serverInfo types.ServerVersionInfo
		err = json.Unmarshal(b, &serverInfo)
		if err != nil {
			data["server_version"] = err.Error()
			output(data)
			log.Fatal("could not unmarshal server response", zap.Error(err))
		}

		data["server_version"] = serverInfo
		output(data)
	},
}
