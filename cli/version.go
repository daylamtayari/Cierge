package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieve program version",
	Run: func(cmd *cobra.Command, args []string) {
		version := getVersion()
		if version != "unknown" {
			version = "version " + version
		}

		fmt.Print(version)
	},
}

// Retrieve the version of the CLI
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		version := info.Main.Version
		if version == "" || version == "(devel)" {
			version = "dev"
		}
		return version
	} else {
		return "unknown"
	}
}
