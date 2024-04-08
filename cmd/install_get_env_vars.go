package cmd

import (
	"strings"

	"github.com/spf13/viper"
)

// read install.pkg.env
// read env.env

func install_get_env_vars(pkg string) map[string]string {
	ans := make(map[string]string)

	strs := []string{}

	if key := pkg + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	} else if key := "install." + pkg + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	} else if key := easifem_current_env_name + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	}

	for _, s := range strs {
		kv := strings.Split(s, "=")
		ans[kv[0]] = kv[1]
	}

	return ans
}
