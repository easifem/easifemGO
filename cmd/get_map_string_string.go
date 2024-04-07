package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

func get_map_string_string(a, b, c string) map[string]string {
	key := a + "." + b + "." + c
	ans := make(map[string]string)

	if viper.IsSet(key) {
		strs := viper.GetStringSlice(key)
		for _, s := range strs {
			kv := strings.Split(s, "=")
			ans[kv[0]] = kv[1]
		}
	}

	return ans
}
