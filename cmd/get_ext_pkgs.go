package cmd

import (
	"github.com/spf13/viper"
)

// this function get a.b.c from the viper
// it uses the default value d
func get_ext_pkgs() []string {
	if key := easifem_current_env_name + ".extpkgs"; viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	return []string{"sparsekit", "lapack95", "fftw", "superlu", "arpack", "tomlf", "lis"}
}
