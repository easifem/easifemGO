/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cleanCmd represents the clean command
// examples
// easifem clean
var cleanCmd = &cobra.Command{
	Use:   "clean pkgname [flags]",
	Short: "A brief description of your command",
	Long:  easifem_clean_intro,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: clean.go | os.Getwd() ➡️ ", err)
		}
		for _, pkg := range args {
			if pkg == "extpkgs" {
				pkgs := get_ext_pkgs()
				for _, p := range pkgs {
					cleanPkgs(p, pwd)
				}
			} else {
				cleanPkgs(pkg, pwd)
			}
		}
	},
}

// clean a package
func cleanPkgs(pkg, pwd string) {
	// source_dir := get_source_dir(pkg)
	build_dir := get_build_dir(pkg)
	// install_dir := get_install_dir(pkg)

	if pwd == build_dir {
		log.Println("[log] :: clean.go() | build_dir is same as current dir")
		log.Println("[log] :: clean.go() | cannot clean current directory for pkg: " + pkg)
	}

	// change_dir(source_dir)
	if err := os.RemoveAll(build_dir); err != nil {
		log.Println("[log] :: clean.go() | could not clean pkg: " + pkg)
	}
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	env := easifem_current_env_name

	cleanCmd.PersistentFlags().StringVarP(&buildDir, "buildDir", "b", easifem_build_dir,
		"Location where easifem will be build, EASIFEM_BUILD_DIR")
	if err := viper.BindPFlag(env+".buildDir",
		cleanCmd.PersistentFlags().Lookup("buildDir")); err != nil {
		log.Fatalln("[err] :: clean.go | viper.BindPFlag() ➡ ", err)
	}

	cleanCmd.PersistentFlags().StringVarP(&sourceDir, "sourceDir", "s", easifem_source_dir,
		"Location where easifem source code will be stored, EASIFEM_SOURCE_DIR")
	if err := viper.BindPFlag(env+".sourceDir",
		cleanCmd.PersistentFlags().Lookup("sourceDir")); err != nil {
		log.Fatalln("[err] :: clean.go | viper.BindPFlag() ➡ ", err)
	}

	cleanCmd.PersistentFlags().StringVarP(&cleanDir, "cleanDir", "i", easifem_clean_dir,
		"Location where easifem will be cleaned, EASIFEM_clean_DIR")
	if err := viper.BindPFlag(env+".cleanDir",
		cleanCmd.PersistentFlags().Lookup("cleanDir")); err != nil {
		log.Fatalln("[err] :: clean.go | viper.BindPFlag() ➡ ", err)
	}

	cleanCmd.PersistentFlags().StringVar(&buildType, "buildType", easifem_build_type,
		"Build type, Release, Debug, Both")
	if err := viper.BindPFlag(env+".buildType",
		cleanCmd.PersistentFlags().Lookup("buildType")); err != nil {
		log.Fatalln("[err] :: viper.BindPFlag() ➡ ", err)
	}

	cleanCmd.PersistentFlags().BoolVar(&buildSharedLibs, "buildSharedLibs", true,
		"Build shared lib")
	if err := viper.BindPFlag(env+".buildSharedLibs",
		cleanCmd.PersistentFlags().Lookup("buildSharedLibs")); err != nil {
		log.Fatalln("[err] :: clean.go | viper.BindPFlag() ➡ ", err)
	}

	cleanCmd.PersistentFlags().BoolVar(&buildStaticLibs, "buildStaticLibs", false,
		"Build Static lib")
	if err := viper.BindPFlag(env+".buildStaticLibs",
		cleanCmd.PersistentFlags().Lookup("buildStaticLibs")); err != nil {
		log.Fatalln("[err] :: clean.go | viper.BindPFlag() ➡ ", err)
	}
}
