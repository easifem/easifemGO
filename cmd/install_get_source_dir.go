package cmd

import (
	"path"

	"github.com/spf13/viper"
)

// this function returns the source directory name
// get install.pkg.sourceDir from the viper
// get sourceDir from the viper
// it uses the default value sourceDir/easifem/pkg
func install_get_source_dir(pkg string) string {
	if key := pkg + ".sourceDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := "install." + pkg + ".sourceDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := easifem_current_env_name + ".sourceDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	return path.Join(sourceDir, "easifem", pkg)
}
