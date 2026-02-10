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
		if info, ok := debug.ReadBuildInfo(); ok {
			version := info.Main.Version
			if version == "" || version == "(devel)" {
				version = "dev"
			}
			fmt.Printf("version %s", version)
		} else {
			fmt.Print("unknown")
		}
	},
}
