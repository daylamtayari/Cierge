package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/daylamtayari/cierge/api"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

const (
	checkmark = `✓`
	crossmark = `✗`
	warnsign  = `⚠`
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get status",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()

		var serverStatus string
		health, err := client.GetHealth()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to fetch server health")
			serverStatus = color.RedString(crossmark + " Server is not reachable")
		} else if health.Server != "cierge" {
			serverStatus = color.YellowString(warnsign + " The specified server host is not a Cierge server")
		} else if health.Status != "ok" {
			serverStatus = color.RedString(crossmark + " Server is in an unhealthy state")
		} else {
			serverStatus = color.GreenString(checkmark + " Server is healthy")
		}

		var userStatus string
		user, err := client.GetMe()
		if err != nil {
			userStatus = color.RedString(crossmark + " Logged out")
			if !errors.Is(err, api.ErrUnauthenticated) {
				logger.Error().Err(err).Msg("Failed to fetch self user")
			}
		} else {
			userStatus = color.GreenString(checkmark + " Logged in - " + user.Email)
		}

		version := getVersion()

		configDir, err := getConfigDir()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve configuration directory")
		}
		configPath := filepath.Join(configDir, "cli.json")

		st := table.NewWriter()
		st.AppendRows([]table.Row{
			{"Server", serverStatus},
			{"User", userStatus},
			{"Version", version},
			{"Config", configPath},
		})
		st.SetStyle(table.StyleLight)
		st.Style().Options.DrawBorder = false
		st.Style().Options.SeparateColumns = false

		fmt.Print(st.Render())
	},
}
