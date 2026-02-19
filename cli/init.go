package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the configuration - run on first use",
	Run: func(cmd *cobra.Command, args []string) {
		if !cmd.Flags().Changed("host") {
			var userHost string
			err := runHuh(huh.NewInput().Title("Enter server URL:").Validate(validateHost).Value(&userHost))
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to prompt user for host")
			}
			cfg.HostURL = userHost
		}

		client := newClient()
		_, err := client.GetHealth()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to get server health")
		}

		err = saveConfig(&cfg)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to save host to config")
		}

		// Log user in
		loginCmd.Run(cmd, args)

		// Display status
		statusCmd.Run(cmd, args)
	},
}

func validateHost(s string) error {
	if s == "" {
		return fmt.Errorf("Host is required") //nolint:staticcheck
	}

	if !strings.Contains(s, "http://") && !strings.Contains(s, "https://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	if u.Host == "" {
		return fmt.Errorf("Hostname must be included") //nolint:staticcheck
	}

	return nil
}
