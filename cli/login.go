package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/charmbracelet/huh"
	"github.com/daylamtayari/cierge/api"
	"github.com/spf13/cobra"
)

var (
	apiKey   string
	username string
	password string

	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to a Cierge instance",
		Run: func(cmd *cobra.Command, args []string) {
			var loginMethod string

			if cmd.Flags().Changed("api-key") {
				loginMethod = "api"
			} else if cmd.Flags().Changed("username") || cmd.Flags().Changed("password") {
				loginMethod = "userpass"
			} else {
				err := huh.NewSelect[string]().
					Title("Select login method").
					Options(
						huh.NewOption("API Key", "api"),
						huh.NewOption("Username/Password", "userpass"),
					).
					Value(&loginMethod).
					Run()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for login method")
				}
			}

			failedCheck := false

			switch loginMethod {
			case "api":
				if apiKey == "" {
					err := huh.NewInput().Title("Enter API key:").EchoMode(huh.EchoModePassword).Value(&apiKey).Run()
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to prompt user for API key")
					}
				}

				client, err := api.NewClient(nil, cfg.HostURL, apiKey)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to create API client")
				}

				_, err = client.GetMe()
				if errors.Is(err, api.ErrUnauthenticated) {
					logger.Fatal().Err(err).Msg("Invalid API key")
				} else if err != nil {
					logger.Error().Err(err).Msg("Failed to verify API key")
					failedCheck = true
				}
				cfg.ApiKey = apiKey

			case "userpass":
				if username == "" {
					err := huh.NewInput().Title("Enter username:").Value(&username).Run()
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to prompt user for username")
					}
				}
				if password == "" {
					err := huh.NewInput().Title("Enter password:").EchoMode(huh.EchoModePassword).Value(&password).Run()
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to prompt user for password")
					}
				}

				client, err := api.NewClient(nil, cfg.HostURL, "")
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to create API client")
				}

				authCookies, err := client.Login(username, password)
				if errors.Is(err, api.ErrUnauthenticated) {
					logger.Fatal().Err(err).Msg("Invalid credentials")
				} else if err != nil {
					logger.Fatal().Err(err).Msg("Failed to authenticate credentials")
				}

				cookieAuthClient, err := createCookieAuthClient(authCookies)
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to create API client")
				}

				user, err := cookieAuthClient.GetMe()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to get user")
				}
				if user.HasApiKey {
					var confirm bool
					err = huh.NewConfirm().Title("You already have an API key, are you sure you want to generate a new one?\nThis will revoke your existing API key.").Value(&confirm).Run()
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to prompt user for confirmation")
					}
					if !confirm {
						// Exit if the user did not confirm overriding the API key
						return
					}
				}

				cfg.ApiKey, err = cookieAuthClient.GenerateAPIKey()
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to generate API key")
				}
			}

			err := saveConfig(&cfg)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to save API key to config")
			} else {
				logger.Debug().Msg("Saved config to file")
			}

			if !failedCheck {
				logger.Info().Msg("Successfully logged in!")
			}
		},
	}
)

func initLoginCmd() *cobra.Command {
	loginCmd.Flags().StringVar(&apiKey, "api-key", "", "API key to use for authentication")
	loginCmd.Flags().StringVar(&username, "username", "", "Username to use for authentication")
	loginCmd.Flags().StringVar(&password, "password", "", "Password to use for authentication")
	loginCmd.MarkFlagsMutuallyExclusive("api-key", "username")
	loginCmd.MarkFlagsMutuallyExclusive("api-key", "password")
	return loginCmd
}

// roundTripperFunc wraps a function to implement http.RoundTripper
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Creates an API client that uses cookie auth
func createCookieAuthClient(authCookies *api.AuthCookies) (*api.Client, error) {
	cookieHeader := fmt.Sprintf("access_token=%s; refresh_token=%s",
		authCookies.AccessToken, authCookies.RefreshToken)

	// HTTP client with custom transport that adds cookies to each request
	cookieClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Cookie", cookieHeader)
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	return api.NewClient(cookieClient, cfg.HostURL, "")
}
