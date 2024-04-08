/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
// examples
// easifem install
var installCmd = &cobra.Command{
	Use:   "install pkgname [flags]",
	Short: "A brief description of your command",
	Long:  easifem_install_intro,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: install.go | os.Getwd() ➡️ ", err)
		}
		for _, pkg := range args {
			if pkg == "extpkgs" {
				pkgs := get_ext_pkgs()
				for _, p := range pkgs {
					if err := installPkgs(p, pwd); err != nil {
						log.Fatalln("[err] :: install.go | installPkgs() ➡ ", err)
					}
				}
			} else {
				if err := installPkgs(pkg, pwd); err != nil {
					log.Fatalln("[err] :: install.go | installPkgs() ➡ ", err)
				}
			}
		}
	},
}

// install a package
func installPkgs(pkg, pwd string) error {
	source_dir := get_source_dir(pkg)
	build_dir := get_build_dir(pkg)
	install_dir := get_install_dir(pkg)

	url, err := get_url("install", pkg)
	if err != nil {
		log.Fatalln("[err] :: install.go |  get_url() ➡ ", err)
	}
	get_pkg(url, source_dir, pwd)
	// change_dir(source_dir)
	make_install_dir(install_dir)

	env_vars := get_map_string_string("install", pkg, "env")
	for k, v := range env_vars {
		fmt.Printf("setting env %s=%s", k, v)
		os.Setenv(k, v)
	}

	switch build_sys := get_build_system("install", pkg); build_sys {
	case "make":
		install_pkg_make(pkg, pwd, source_dir, build_dir, install_dir,
			get_string_slice_value("install", pkg, "buildOptions"))
	case "cmake":
		install_pkg_cmake(pkg, pwd, source_dir, build_dir, install_dir,
			install_get_string_value(pkg, "buildType", buildType),
			install_get_bool_value(pkg, "buildSharedLibs", buildSharedLibs),
			install_get_bool_value(pkg, "buildStaticLibs", buildStaticLibs),
			install_get_string_slice_value(pkg, "buildOptions"))
	case "mason":
		install_pkg_mason(pkg, pwd, source_dir, build_dir, install_dir)
	default:
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
