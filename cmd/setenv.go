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

// setenvCmd represents the setenv command
var setenvCmd = &cobra.Command{
	Use:   "setenv",
	Short: "Set environment variables for running easifem on your system.",
	Long:  easifem_setenv_intro,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(easifem_banner)
		if err := cmd.Help(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		buildDir = viper.GetString(easifem_current_env_name + ".buildDir")
		sourceDir = viper.GetString(easifem_current_env_name + ".sourceDir")
		installDir = viper.GetString(easifem_current_env_name + ".installDir")

		showConfig()
	},
}

func init() {
	rootCmd.AddCommand(setenvCmd)

	setenvCmd.PersistentFlags().StringVarP(&buildDir, "buildDir", "b", easifem_build_dir,
		"Location where easifem will be build, EASIFEM_BUILD_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".buildDir",
		setenvCmd.PersistentFlags().Lookup("buildDir")); err != nil {
		log.Println("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
		os.Exit(1)
	}

	setenvCmd.PersistentFlags().StringVarP(&sourceDir, "sourceDir", "s", easifem_source_dir,
		"Location where easifem source code will be stored, EASIFEM_SOURCE_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".sourceDir",
		setenvCmd.PersistentFlags().Lookup("sourceDir")); err != nil {
		log.Println("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
		os.Exit(1)
	}

	setenvCmd.PersistentFlags().StringVarP(&installDir, "installDir", "i", easifem_install_dir,
		"Location where easifem will be installed, EASIFEM_INSTALL_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".installDir",
		setenvCmd.PersistentFlags().Lookup("installDir")); err != nil {
		log.Println("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
		os.Exit(1)
	}
}
