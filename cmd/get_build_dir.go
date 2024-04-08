package cmd

import (
	"path"

	"github.com/spf13/viper"
)

// this function returns the build directory name
// get install.pkg.buildDir from the viper
// get buildDir from the viper
// it uses the default value buildDir/easifem/pkg
func get_build_dir(pkg string) string {
	if key := "install." + pkg + ".buildDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := easifem_current_env_name + ".buildDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	return path.Join(buildDir, "easifem", pkg)
}
