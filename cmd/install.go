/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		// fmt.Println(easifem_banner)
		// if err := cmd.Help(); err != nil {
		// 	log.Println(err)
		// 	os.Exit(1)
		// }
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[INTERNAL ERROR] :: install.go | os.Getwd() ➡️ ", err)
		}
		for _, pkg := range args {
			if err := installPkgs(pkg, pwd); err != nil {
				log.Fatalln("[INTERNAL ERROR] :: install.go | installPkgs() ➡ ", err)
			}
		}
	},
}

// install a package
func installPkgs(pkg, pwd string) error {
	source_dir := get_string_value("install", pkg, "sourceDir", sourceDir)
	build_dir := get_string_value("install", pkg, "buildDir", buildDir)
	install_dir := get_string_value("install", pkg, "installDir", installDir)
	url, err := get_url("install", pkg)
	if err != nil {
		log.Fatalln("[INTERNAL ERROR] :: install.go |  get_url() ➡ ", err)
	}
	get_pkg(url, source_dir, pwd)
	change_dir(source_dir)
	make_install_dir(install_dir)

	switch build_sys := get_build_system("install", pkg); build_sys {
	case "make":
		install_pkg_make(pkg, pwd, source_dir, build_dir, install_dir)
	case "cmake":
		install_pkg_cmake(pkg, pwd, source_dir, build_dir, install_dir,
			get_string_value("install", pkg, "buildType", buildType),
			get_bool_value("install", pkg, "buildSharedLibs", buildSharedLibs),
			get_bool_value("install", pkg, "buildStaticLibs", buildStaticLibs),
			get_string_slice_value("install", pkg, "buildOptions"))
	case "mason":
		install_pkg_mason(pkg, pwd, source_dir, build_dir, install_dir)
	default:
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)

	env := easifem_current_env_name

	installCmd.PersistentFlags().StringVarP(&buildDir, "buildDir", "b", easifem_build_dir,
		"Location where easifem will be build, EASIFEM_BUILD_DIR")
	if err := viper.BindPFlag(env+".buildDir",
		installCmd.PersistentFlags().Lookup("buildDir")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: install.go | viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().StringVarP(&sourceDir, "sourceDir", "s", easifem_source_dir,
		"Location where easifem source code will be stored, EASIFEM_SOURCE_DIR")
	if err := viper.BindPFlag(env+".sourceDir",
		installCmd.PersistentFlags().Lookup("sourceDir")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: install.go | viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().StringVarP(&installDir, "installDir", "i", easifem_install_dir,
		"Location where easifem will be installed, EASIFEM_INSTALL_DIR")
	if err := viper.BindPFlag(env+".installDir",
		installCmd.PersistentFlags().Lookup("installDir")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: install.go | viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().StringVar(&buildType, "buildType", easifem_build_type,
		"Build type, Release, Debug, Both")
	if err := viper.BindPFlag(env+".buildType",
		installCmd.PersistentFlags().Lookup("buildType")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().BoolVar(&buildSharedLibs, "buildSharedLibs", true,
		"Build shared lib")
	if err := viper.BindPFlag(env+".buildSharedLibs",
		installCmd.PersistentFlags().Lookup("buildSharedLibs")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: install.go | viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().BoolVar(&buildStaticLibs, "buildStaticLibs", false,
		"Build Static lib")
	if err := viper.BindPFlag(env+".buildStaticLibs",
		installCmd.PersistentFlags().Lookup("buildStaticLibs")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: install.go | viper.BindPFlag() ➡ ", err)
	}
}
