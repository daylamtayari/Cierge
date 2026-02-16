package main

import "github.com/spf13/cobra"

var (
	platform string

	tokenCmd = &cobra.Command{
		Use:   "token",
		Short: "Manage platform tokens",
	}
)

func initTokenCmd() *cobra.Command {
	tokenListCmd.PersistentFlags().StringVarP(&platform, "platform", "p", "", "Platform to get tokens for")
	tokenCmd.AddCommand(tokenAddCmd)
	tokenCmd.AddCommand(initTokenListCmd())
	return tokenCmd
}
