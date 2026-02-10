package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	HostURL string `json:"host_url" mapstructure:"host_url"`
	ApiKey  string `json:"api_key" mapstructure:"api_key"`
}

// Initializes the configuration using Viper
func initConfig() (*config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	viper.SetConfigName("cli")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("CIERGE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create if it does not exist
			if err := saveConfig(&config{}); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// Enforce secure permissions on config file
		if err := checkConfigPermissions(viper.ConfigFileUsed()); err != nil {
			return nil, err
		}
	}

	var cfg config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("host_url", "http://localhost:8080")
}

func saveConfig(cfg *config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "cli.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	return nil
}

func getConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "cierge"), nil
}

// checkConfigPermissions validates that the config file has secure permissions (0600)
func checkConfigPermissions(configPath string) error {
	info, err := os.Stat(configPath)
	if err != nil {
		return err
	}

	// Require 0600 permissions
	if info.Mode().Perm() != 0600 {
		return fmt.Errorf("insecure config file permissions (%v). Run: chmod 0600 %s",
			info.Mode().Perm(), configPath)
	}

	return nil
}
