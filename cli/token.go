package main

import "github.com/spf13/cobra"

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage platform tokens",
}

func initTokenCmd() *cobra.Command {
	tokenCmd.AddCommand(initTokenListCmd())
	return tokenCmd
}
