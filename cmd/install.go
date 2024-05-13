/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package cmd

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
	Short: "A brief description of your command",
	Long:  easifem_install_intro,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: install.go | os.Getwd() ➡️ ", err)
		}

		for _, pkg := range args {
			var err error
			if pkg == "extpkgs" {
				err = installExtPkgs(pwd)
			} else {
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
	var p *Pkg
	var err error

	if !quiet {
		log.Println("[log] :: install.go | installPkgs() | pkg ➡️ ", pkg)
	}

	p, err = PkgMakeFromToml(pkg)
	if err != nil {
		return fmt.Errorf("installPkgs() | pkg=%s, err=%w", pkg, err)
	}

	err = PkgInstall(p, pwd)
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
	pkgs := pkgGetExtPkgs()
	var p *Pkg
	var err error

	for _, pkg := range pkgs {

		p, err = PkgMakeFromToml(pkg)
		if err != nil {
			return fmt.Errorf("installExtPkgs() | pkg=%s, err=%w", pkg, err)
		}

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
//                                                                      init
//----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(installCmd)
}
