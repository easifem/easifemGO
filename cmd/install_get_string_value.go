package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

// this function get a string from install.pkg.key
// this function get a string from env.key
// it uses the default value d
func install_get_string_value(pkg, key, val string) string {
	if key := strings.Join([]string{pkg, key}, "."); viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := strings.Join([]string{"install", pkg, key}, "."); viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := strings.Join([]string{easifem_current_env_name, key}, "."); viper.IsSet(key) {
		return viper.GetString(key)
	}
	return val
}
