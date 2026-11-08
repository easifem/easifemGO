/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
// examples
// easifem clean
var cleanCmd = &cobra.Command{
	Use:   "clean pkgname [flags]",
	Short: "Remove package build and install directories; reject paths containing the current directory",
	Long:  easifem_clean_intro,
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		candidates := pkgGetAllNames()
		candidates = append(candidates, "extpkgs")
		return candidates, cobra.ShellCompDirectiveDefault
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
					if err := cleanPkgs(p, pwd); err != nil {
						log.Fatal(err)
					}
				}
			} else {
				if err := cleanPkgs(pkg, pwd); err != nil {
					log.Fatal(err)
				}
			}
		}
	},
}

// clean a package
func cleanPkgs(name, pwd string) error {
	pkg, ok := easifem_pkgs[name]
	if !ok || pkg == nil {
		return fmt.Errorf("unknown package %q", name)
	}
	// Validate both paths before removing either directory.
	for _, dir := range []string{pkg.BuildDir, pkg.InstallDir} {
		if err := validateCleanPath(dir, pwd); err != nil {
			return err
		}
	}
	for _, dir := range []string{pkg.BuildDir, pkg.InstallDir} {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("clean %s: %w", dir, err)
		}
	}
	return nil
}

func validateCleanPath(dir, pwd string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("refusing to clean an empty path")
	}
	target, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	// Resolve existing symlinks to catch aliases of the current directory.
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		target = resolved
	} else if !os.IsNotExist(err) {
		return err
	}
	current, err := filepath.Abs(pwd)
	if err != nil {
		return err
	}
	current, err = filepath.EvalSymlinks(current)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(target, current)
	if err != nil {
		return err
	}
	if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return fmt.Errorf("refusing to clean %s: contains the current directory", dir)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
