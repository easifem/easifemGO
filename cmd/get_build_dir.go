package cmd

import (
	"path"

	"github.com/spf13/viper"
)

// this function get a.b.c from the viper
// it uses the default value d
func get_build_dir(pkg string) string {
	if key := "install." + pkg + ".buildDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	return path.Join(buildDir, "easifem", pkg)
}
