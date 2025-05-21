/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package internal

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
// examples
// easifem clean
var cleanCmd = &cobra.Command{
	Use:   "clean pkgname [flags]",
	Short: "remove build and install files of a pkg",
	Long:  easifem_clean_intro,
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return (pkgGetAllNames()), (cobra.ShellCompDirectiveDefault)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: clean.go | os.Getwd() ➡️ ", err)
		}
		for _, pkg := range args {
			if pkg == "extpkgs" {
				pkgs := pkgGetExtPkgs()
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
	// source_dir := install_get_source_dir(pkg)
	build_dir := easifem_pkgs[pkg].BuildDir
	install_dir := easifem_pkgs[pkg].InstallDir

	if pwd == build_dir {
		log.Println("[log] :: clean.go() | build_dir is same as current dir")
		log.Println("[log] :: clean.go() | cannot clean current directory for pkg: " + pkg)
	}

	if err := os.RemoveAll(build_dir); err != nil {
		log.Println("[log] :: clean.go() | could not clean pkg: " + pkg)
	}

	if pwd == install_dir {
		log.Println("[log] :: clean.go() | install_dir is same as current dir")
		log.Println("[log] :: clean.go() | cannot clean current directory for pkg: " + pkg)
	}

	if err := os.RemoveAll(install_dir); err != nil {
		log.Println("[log] :: clean.go() | could not clean pkg: " + pkg)
	}
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
