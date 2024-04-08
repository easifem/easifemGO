package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

// this function get a bool from install.pkg.key
// this function get a bool from env.key
// it uses the default value d
func install_get_bool_value(pkg, key string, val bool) bool {
	if key := strings.Join([]string{"install", pkg, key}, "."); viper.IsSet(key) {
		return viper.GetBool(key)
	}
	if key := strings.Join([]string{easifem_current_env_name, key}, "."); viper.IsSet(key) {
		return viper.GetBool(key)
	}
	return val
}
