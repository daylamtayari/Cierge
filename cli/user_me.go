package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Information about your user",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()

		ut := table.NewWriter()
		ut.SetStyle(table.StyleLight)
		ut.Style().Options.DrawBorder = false
		ut.Style().Options.SeparateColumns = false

		user, err := client.GetMe()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to get user")
		}

		ut.AppendRows([]table.Row{
			{"ID", user.ID},
			{"Email", user.Email},
			{"Has API Key", user.HasApiKey},
			{"Admin", user.IsAdmin},
			{"Created At", user.CreatedAt.Local().Format("2006-01-02 15:04:05 MST")},
		})

		fmt.Print(ut.Render() + "\n")
	},
}
