package cmd

import "github.com/spf13/viper"

// this function get a.b.c from the viper
// it uses the default value d
func get_string_value(a, b, c, d string) string {
	var ans string
	key := a + "." + b + "." + c
	if viper.IsSet(key) {
		ans = viper.GetString(key)
	} else {
		ans = d
	}
	return ans
}
