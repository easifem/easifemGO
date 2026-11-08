/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package internal

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
// examples
// easifem install
var installCmd = &cobra.Command{
	Use:   "install pkgname [flags]",
	Short: "This subcommand install one or more packages including dependencies.",
	Long:  easifem_install_intro,
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		candidates := pkgGetAllNames()
		candidates = append(candidates, "easifem", "extpkgs")
		return candidates, cobra.ShellCompDirectiveDefault
	},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		for _, pkg := range args {
			var err error
			switch name := pkg; name {
			case "extpkgs":
				err = installExtPkgs(pwd)
			case "easifem":
				err = installEasifem(pwd)
			default:
				err = installPkgs(pkg, pwd)
			}
			if err != nil {
				return err
			}
		}
		return nil
	},
}

//----------------------------------------------------------------------------
//                                                            installPkgs
//----------------------------------------------------------------------------

// install a single package
func installPkgs(pkg, pwd string) error {
	var err error

	// check if pkg is in the list of easifem_pkgs
	if _, ok := easifem_pkgs[pkg]; !ok {
		return fmt.Errorf("unknown package %q", pkg)
	}

	started := time.Now()
	err = PkgInstall(easifem_pkgs[pkg], pwd)
	if err != nil {
		return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
	}

	if easifem_pkgs[pkg].IsActive && !quiet {
		fmt.Fprintf(os.Stderr, "\n  ✓ %s installed successfully · %s\n    %s\n", pkg, time.Since(started).Round(100*time.Millisecond), easifem_pkgs[pkg].InstallDir)
	}
	return nil
}

//----------------------------------------------------------------------------
//                                                          installExtPkgs
//----------------------------------------------------------------------------

// install a package
func installExtPkgs(pwd string) error {
	var names []string
	for name, pkg := range easifem_pkgs {
		if pkg.IsExtPkg {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if err := installPkgs(name, pwd); err != nil {
			return err
		}
	}
	return nil
}

//----------------------------------------------------------------------------
//                                                          installEasifem
//----------------------------------------------------------------------------

// install a package
func installEasifem(pwd string) error {
	return installEasifemWith(pwd, installPkgs)
}

func installEasifemWith(pwd string, install func(string, string) error) error {
	var err error

	for _, name := range []string{"base", "classes"} {
		if _, ok := easifem_pkgs[name]; !ok {
			continue
		}
		if err = install(name, pwd); err != nil {
			return err
		}
	}

	return err
}

//----------------------------------------------------------------------------
//                                                                      init
//----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(installCmd)
}
