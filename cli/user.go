package main

import "github.com/spf13/cobra"

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage user",
}

func initUserCmd() *cobra.Command {
	userCmd.AddCommand(userMeCmd)
	return userCmd
}
