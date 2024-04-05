package cmd

import (
	"github.com/spf13/viper"
)

// this function returns the url from config file
func get_build_system(a, b string) string {
	if key := a + "." + b + "." + "buildSystem"; viper.IsSet(key) {
		return viper.GetString(key)
	} else {
		return "cmake"
	}
}
