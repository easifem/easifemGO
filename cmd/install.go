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
		fmt.Println("install called args, ", args)

		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[INTERNAL ERROR] :: os.Getwd() ➡️ ", err)
		} else {
			log.Printf("pwd = %s", pwd)
		}

		for _, pkg := range args {
			if err := installPkgs(pkg, pwd); err != nil {
				log.Fatalln("[INTERNAL ERROR] :: installPkgs() ➡ ", err)
			}
		}
	},
}

// install a package
func installPkgs(pkg, pwd string) error {
	source_dir := get_string_value("install", pkg, "sourceDir", sourceDir)
	build_dir := get_string_value("install", pkg, "buildDir", buildDir)
	install_dir := get_string_value("install", pkg, "installDir", installDir)
	log.Printf("pkg %s will store at ➡ %s\n", pkg, source_dir)
	log.Printf("pkg %s will build at ➡ %s\n", pkg, build_dir)
	log.Printf("pkg %s will install at ➡ %s\n", pkg, install_dir)
	log.Printf("pkg %s downloading ... \n", pkg)

	url, err := get_url("install", pkg)
	if err != nil {
		log.Fatalln("[INTERNAL ERROR] :: get_url() ➡ ", err)
	}
	get_pkg(url, source_dir, pwd)
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)

	env := easifem_current_env_name

	installCmd.PersistentFlags().StringVarP(&buildDir, "buildDir", "b", easifem_build_dir,
		"Location where easifem will be build, EASIFEM_BUILD_DIR")
	if err := viper.BindPFlag(env+".buildDir",
		installCmd.PersistentFlags().Lookup("buildDir")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().StringVarP(&sourceDir, "sourceDir", "s", easifem_source_dir,
		"Location where easifem source code will be stored, EASIFEM_SOURCE_DIR")
	if err := viper.BindPFlag(env+".sourceDir",
		installCmd.PersistentFlags().Lookup("sourceDir")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
	}

	installCmd.PersistentFlags().StringVarP(&installDir, "installDir", "i", easifem_install_dir,
		"Location where easifem will be installed, EASIFEM_INSTALL_DIR")
	if err := viper.BindPFlag(env+".installDir",
		installCmd.PersistentFlags().Lookup("installDir")); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
	}
}
