package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

func install_get_string_slice_value(pkg, key string) []string {
	if key := strings.Join([]string{pkg, key}, "."); viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	if key := strings.Join([]string{"install", pkg, key}, "."); viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	return nil
}
