/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package internal

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
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: install.go | os.Getwd() ➡️ ", err)
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
				log.Fatalln("[err] :: install.go | installPkgs() ➡ ", err)
			}
		}
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
		return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
	}

	err = PkgInstall(easifem_pkgs[pkg], pwd)
	if err != nil {
		return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
	}

	return err
}

//----------------------------------------------------------------------------
//                                                          installExtPkgs
//----------------------------------------------------------------------------

// install a package
func installExtPkgs(pwd string) error {
	var err error

	for pkg, p := range easifem_pkgs {
		if p.IsExtPkg {
			if !quiet {
				log.Println("[log] :: install.go | installExtPkgs() | pkg ➡️ ", pkg)
			}

			err = PkgInstall(p, pwd)
			if err != nil {
				return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
			}
		}
	}

	return err
}

//----------------------------------------------------------------------------
//                                                          installEasifem
//----------------------------------------------------------------------------

// install a package
func installEasifem(pwd string) error {
	var err error

	for pkg, p := range easifem_pkgs {
		switch name := pkg; name {
		case "base":
			if !quiet {
				log.Println("[log] :: install.go | installEasifem() | pkg ➡️ ", pkg)
			}

			err = PkgInstall(p, pwd)
			if err != nil {
				return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
			}
		case "classes":
			if !quiet {
				log.Println("[log] :: install.go | installEasifem() | pkg ➡️ ", pkg)
			}

			err = PkgInstall(p, pwd)
			if err != nil {
				return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
			}
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
