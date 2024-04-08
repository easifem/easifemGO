package cmd

import (
	"github.com/spf13/viper"
)

// this function returns the url from config file
func install_get_build_system(pkg string) string {
	if key := pkg + ".buildSystem"; viper.IsSet(key) {
		return viper.GetString(key)
	}

	if key := "install." + pkg + ".buildSystem"; viper.IsSet(key) {
		return viper.GetString(key)
	}

	if key := easifem_current_env_name + ".buildSystem"; viper.IsSet(key) {
		return viper.GetString(key)
	}

	return "cmake"
}
