package config

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// setDefaults walks the struct and registers the defaults using the tags,
// avoiding manually setting the values for every item with viper.SetDefault
func setDefaults(v *viper.Viper) {
	setDefaultsFromTags(v, reflect.TypeFor[Config](), "")
}

func setDefaultsFromTags(v *viper.Viper, t reflect.Type, prefix string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Key path building
		key := jsonTag
		if prefix != "" {
			key = prefix + "." + jsonTag
		}

		// Nested struct handling
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeFor[time.Time]() {
			setDefaultsFromTags(v, field.Type, key)
			continue
		}

		defaultVal := field.Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		v.SetDefault(key, parseDefault(field.Type, defaultVal))
	}
}

func parseDefault(t reflect.Type, val string) any {
	switch t.Kind() {
	case reflect.String:
		return val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle duration type
		if t == reflect.TypeFor[Duration]() {
			d, _ := time.ParseDuration(val)
			return d
		}
		i, _ := strconv.ParseInt(val, 10, 64)
		return int(i)
	case reflect.Bool:
		return val == "true"
	case reflect.Slice:
		if val == "" {
			return []string{}
		}
		return strings.Split(val, ",")
	default:
		return val
	}
}
