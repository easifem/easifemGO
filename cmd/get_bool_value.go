package cmd

import "github.com/spf13/viper"

// this function get a.b.c from the viper
// it uses the default value d
func get_bool_value(a, b, c string, d bool) bool {
	key := a + "." + b + "." + c
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	return d
}
