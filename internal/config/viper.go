package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var v *viper.Viper

func V() *viper.Viper {
	if v == nil {
		v = viper.New()
		setDefaults(v)
	}
	return v
}

func Load() (*Config, error) {
	v := V()

	v.SetConfigName("config")
	v.SetConfigType("json")

	// Config paths supporting same dir and ~/.config dir
	v.AddConfigPath(".")
	v.AddConfigPath("~/.config/cierge")
	v.AddConfigPath("/etc/cierge")

	// Environment variables
	v.SetEnvPrefix("CIERGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	//
	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// No return as could be specified by env variables and defaults
	}

	var cfg Config
	if err := unmarshalConfig(&cfg); err != nil {
		return nil, err
	}

	// Validate the config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Unmarshal config using the json tags instead of having duplicate mapstructure tags
func unmarshalConfig(cfg *Config) error {
	if err := v.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	}); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
