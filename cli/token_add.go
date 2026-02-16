package main

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daylamtayari/cierge/resy"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	tokenAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Connect a reservation platform",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()

			if cmd.Flags().Changed("platform") {
				platform = strings.ToLower(platform)
				if platform != "resy" && platform != "opentable" {
					logger.Fatal().Msgf("Invalid platform %q specified - only 'resy' and 'opentable' are supported platforms", platform)
				}
			} else {
				err := huh.NewSelect[string]().
					Title("Select reservation platform:").
					Options(
						huh.NewOption("Resy", "resy"),
						huh.NewOption("OpenTable", "opentable"),
					).
					Value(&platform).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for reservation platform")
				}
			}

			var token any
			switch platform {
			case "resy":
				var username string
				var password string

				err := huh.NewInput().Title("Enter your email:").
					Description(color.YellowString(warnsign + " By connecting your credentials, you assume trust in the server owner")).
					Value(&username).Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for username")
				}

				err = huh.NewInput().Title("Enter your password:").
					Description(color.YellowString(warnsign + " By connecting your credentials, you assume trust in the server owner")).
					EchoMode(huh.EchoModePassword).Value(&password).Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for password")
				}

				resyClient := resy.NewClient(nil, resy.Tokens{ApiKey: resy.DefaultApiKey}, "")
				resyToken, err := resyClient.Login(username, password)
				if err != nil && errors.Is(err, resy.ErrUnauthorized) {
					logger.Fatal().Msg("Invalid Resy credentials provided")
				} else if err != nil {
					logger.Fatal().Err(err).Msg("Failed to login")
				}

				token = resyToken

			case "opentable":
				// TODO: Complete opentable implementation
			}

			_, err := client.CreatePlatformToken(platform, token)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to create platform token")
			}

			logger.Info().Msgf("Successfully created platform token for platform %s", platform)
		},
	}
)
