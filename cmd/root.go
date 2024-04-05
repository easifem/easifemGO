/*
Copyright © 2024 Vikas Sharma, Ph.D. vickysharma0812@gmail.com
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "easifem",
	Short: "A brief description of your application",
	Long:  easifem_intro,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(easifem_banner)
		if err := cmd.Help(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		showConfig()
	},
}

func showConfig() {
	fmt.Println("configFile: ", configFile)
	fmt.Println("envName: ", easifem_current_env_name)
	fmt.Println("quiet: ", quiet)
	fmt.Println("buildDir: ", buildDir)
	fmt.Println("sourceDir: ", sourceDir)
	fmt.Println("installDir: ", installDir)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "",
		"Config file name with extension (e.g. easifem.toml)")

	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Run commands in quiet mode.")
	if err := viper.BindPFlag(easifem_current_env_name+".quiet",
		rootCmd.PersistentFlags().Lookup("quiet")); err != nil {
		log.Println("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&easifem_current_env_name,
		"env", easifem_default_env_name, "Current environment name")
	if err := viper.BindPFlag("envName",
		rootCmd.PersistentFlags().Lookup("env")); err != nil {
		log.Println("[INTERNAL ERROR] :: viper.BindPFlag() ➡ ", err)
		os.Exit(1)
	}
}

func initConfig() {
	if configFile == "" {
		viper.SetConfigName(easifem_config_name)
		viper.SetConfigType(easifem_config_type)

		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Println("[INTERNAL ERROR] :: homedir.Dir() ➡ ", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(path.Join(home, easifem_config_dir))
		viper.AddConfigPath(path.Join(home, ".config"))
		viper.AddConfigPath(path.Join(home, ".config", easifem_config_dir))
		viper.AddConfigPath(".")

	}

	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err == nil {
		quiet = viper.GetBool(easifem_current_env_name + ".quiet")
		easifem_current_env_name = viper.GetString("envName")
		configFile = viper.ConfigFileUsed()
		log.Println("[log] :: Success in reading config file ➡️ " + viper.ConfigFileUsed())
	} else {
		log.Println("[INTERNAL ERROR] :: viper.ReadInConfig() ➡ ", err)
		os.Exit(1)
	}
}
