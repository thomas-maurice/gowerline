package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns the version of the binary",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		output(types.ServerVersionInfo{
			BuildHost:       version.BuildHost,
			BuildTime:       version.BuildTime,
			GitHash:         version.BuildHash,
			Version:         version.Version,
			OperatingSystem: version.OS,
			Architecture:    version.Arch,
		})
	},
}
