/*
Copyright © 2024 Vikas Sharma, Ph.D. vickysharma0812@gmail.com
*/

package internal

import (
	"fmt"
	"io"
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
			log.Fatalln("[err] :: root.go | cmd.Help() ➡️ ", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln("[err] :: root.go | Execute() ➡️ ", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "",
		"Config file name with extension (e.g. easifem.toml)")

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false,
		"Run commands in quiet mode.")

	rootCmd.PersistentFlags().BoolVar(&noDownload, "no-download", false,
		"download packages")

	rootCmd.PersistentFlags().StringVar(&easifem_current_env_name,
		"env", easifem_default_env_name, "Current environment name")
	if err := viper.BindPFlag("envName",
		rootCmd.PersistentFlags().Lookup("env")); err != nil {
		log.Fatalln("[err] :: root.go | viper.BindPFlag() ➡ ", err)
	}

	setenvCmd.PersistentFlags().StringVarP(&buildDir, "buildDir", "b", easifem_build_dir,
		"Location where easifem will be build, EASIFEM_BUILD_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".buildDir",
		setenvCmd.PersistentFlags().Lookup("buildDir")); err != nil {
		log.Fatalln("[err] :: viper.BindPFlag() ➡ ", err)
	}

	setenvCmd.PersistentFlags().StringVarP(&sourceDir, "sourceDir", "s", easifem_source_dir,
		"Location where easifem source code will be stored, EASIFEM_SOURCE_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".sourceDir",
		setenvCmd.PersistentFlags().Lookup("sourceDir")); err != nil {
		log.Fatalln("[err] :: viper.BindPFlag() ➡ ", err)
	}

	setenvCmd.PersistentFlags().StringVarP(&installDir, "installDir", "i", easifem_install_dir,
		"Location where easifem will be installed, EASIFEM_INSTALL_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".installDir",
		setenvCmd.PersistentFlags().Lookup("installDir")); err != nil {
		log.Fatalln("[err] :: viper.BindPFlag() ➡ ", err)
	}

	setenvCmd.PersistentFlags().StringVarP(&lintDir, "lintDir", "l", easifem_lint_dir,
		"Location where easifem will be linted, EASIFEM_lint_DIR")
	if err := viper.BindPFlag(easifem_current_env_name+".lintDir",
		setenvCmd.PersistentFlags().Lookup("lintDir")); err != nil {
		log.Fatalln("[err] :: viper.BindPFlag() ➡ ", err)
	}
}

func initConfig() {
	if quiet {
		log.SetOutput(io.Discard)
	}

	if configFile == "" {
		viper.SetConfigName(easifem_config_name)
		viper.SetConfigType(easifem_config_type)

		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalln("[err] :: root.go | homedir.Dir() ➡ ", err)
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
		easifem_current_env_name = viper.GetString("envName")
		// quiet = viper.GetBool(easifem_current_env_name + ".quiet")
		configFile = viper.ConfigFileUsed()
		configPath = path.Dir(configFile)
	} else {
		log.Fatalln("[err] :: root.go | viper.ReadInConfig() ➡ ", err)
	}

	buildDir = viper.GetString(easifem_current_env_name + ".buildDir")
	sourceDir = viper.GetString(easifem_current_env_name + ".sourceDir")
	installDir = viper.GetString(easifem_current_env_name + ".installDir")
	lintDir = viper.GetString(easifem_current_env_name + ".lintDir")

	buildDir = os.ExpandEnv(buildDir)
	sourceDir = os.ExpandEnv(sourceDir)
	installDir = os.ExpandEnv(installDir)
	lintDir = os.ExpandEnv(lintDir)
	pkgMakeDir(lintDir)

	if err := makeAllPkgsFromToml(); err != nil {
		log.Fatalln("[err] :: root.go | initConfig() ➡ ", err)
	}

	buildCache()

	if err := writeCache(); err != nil {
		log.Fatalln("[err] :: root.go | initConfig() ➡ ", err)
	}

	writeShellVarFle()
}
