package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/rs/zerolog"
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

	v.SetConfigName("server")
	v.SetConfigType("json")

	// Config paths supporting same dir and ~/.config dir
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	v.AddConfigPath(".")
	v.AddConfigPath(filepath.Join(userConfigDir, "cierge"))
	v.AddConfigPath("/etc/cierge")

	// Environment variables
	v.SetEnvPrefix("CIERGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// No configuration file found but could be specified by env variables and defaults
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
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			stringToZerologLevelHook(),
			stringToDurationHook(),
		)
	}); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// Decode hook for zerolog.Level for log level value
func stringToZerologLevelHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeFor[zerolog.Level]() {
			return data, nil
		}

		str := data.(string)
		level, err := zerolog.ParseLevel(str)
		if err != nil {
			return nil, fmt.Errorf("invalid log level %q: %w", str, err)
		}
		return level, nil
	}
}

// Decode hook for the duration type
func stringToDurationHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeFor[Duration]() {
			return data, nil
		}

		str := data.(string)
		duration, err := time.ParseDuration(str)
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q: %w", str, err)
		}
		return Duration(duration), nil
	}
}
