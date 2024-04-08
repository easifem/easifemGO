package cmd

import (
	"path"

	"github.com/spf13/viper"
)

// this function returns the install directory name
// get install.pkg.installDir from the viper
// get installDir from the viper
// it uses the default value installDir/easifem/pkg
func get_install_dir(pkg string) string {
	if key := "install." + pkg + ".installDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := easifem_current_env_name + ".installDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	return path.Join(installDir, "easifem", pkg)
}
