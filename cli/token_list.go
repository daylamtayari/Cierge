package main

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	tokenListCmd = &cobra.Command{
		Use:   "list",
		Short: "List platform tokens",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()

			if cmd.Flags().Changed("platform") {
				platform = strings.ToLower(platform)
				if platform != "resy" && platform != "opentable" {
					logger.Fatal().Msg("Invalid platform %q specified - only 'resy' and 'opentable' are valid platforms")
				}
			}

			tokens, err := client.GetPlatformTokens(&platform)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to retrieve platform tokens")
			}

			tt := table.NewWriter()
			tt.SetStyle(table.StyleRounded)
			tt.AppendHeader(table.Row{"ID", "Platform", "Has Refresh", "Expires At", "Refresh Expires At", "Created At"})

			for _, token := range tokens {
				tt.AppendRow(table.Row{
					token.ID,
					token.Platform,
					token.HasRefresh,
					token.ExpiresAt.Local().Format("2006-01-02 15:04:05"),
					token.RefreshExpiresAt.Local().Format("2006-01-02 15:04:05"),
					token.CreatedAt.Local().Format("2006-01-02 15:04:05"),
				})
			}

			fmt.Print(tt.Render() + "\n")
		},
	}
)

func initTokenListCmd() *cobra.Command {
	return tokenListCmd
}
