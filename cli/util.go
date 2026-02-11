package main

import "github.com/daylamtayari/cierge/api"

// Wraps api.NewClient to create a new client
// using the values from the config file
// and handles any error
func newClient() *api.Client {
	client, err := api.NewClient(nil, cfg.HostURL, cfg.ApiKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create API client")
	}
	return client
}
