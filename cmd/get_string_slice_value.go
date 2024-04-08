package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

func get_string_slice_value(a, b, c string) []string {
	key := a + "." + b + "." + c

	if viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}

	return nil
}

func install_get_string_slice_value(pkg, key string) []string {
	if key := strings.Join([]string{"install", pkg, key}, "."); viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	return nil
}
