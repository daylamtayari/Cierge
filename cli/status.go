package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

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

		st := table.NewWriter()
		st.SetStyle(table.StyleLight)
		st.Style().Options.DrawBorder = false
		st.Style().Options.SeparateColumns = false

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
		st.AppendRow(table.Row{"Server", serverStatus})

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
		st.AppendRow(table.Row{"User", userStatus})

		resyStatus := "Unknown"
		openTableStatus := "Unknown"
		if user != nil {
			platformTokens, err := client.GetPlatformTokens(nil)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to fetch platform tokens")
			} else {
				resyStatus = color.RedString(crossmark + " Not connected")
				openTableStatus = color.RedString(crossmark + " Not connected")
				for _, token := range platformTokens {
					var tokenStatus string
					if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
						tokenStatus = color.YellowString(warnsign + " Token is expired")
					} else {
						tokenStatus = color.GreenString(checkmark + " Connected")
					}

					switch token.Platform {
					case "resy":
						resyStatus = tokenStatus
					case "opentable":
						openTableStatus = tokenStatus
					}
				}
			}
		}
		st.AppendRows([]table.Row{
			{"Resy", resyStatus},
			{"OpenTable", openTableStatus},
		})

		st.AppendRow(table.Row{"Version", getVersion()})

		configDir, err := getConfigDir()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve configuration directory")
		}
		st.AppendRow(table.Row{"Config", filepath.Join(configDir, "cli.json")})

		fmt.Print(st.Render() + "\n")
	},
}
