/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

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

			if pkg == "extpkgs" {
				if err := installExtPkgs(pwd); err != nil {
					log.Fatalln("[err] :: install.go | installExtPkgs() ➡ ", err)
				}
			}

			if err := installPkgs(pkg, pwd); err != nil {
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
	p, err := PkgMakeFromToml(path.Join(configPath, pkgConfigDir, pkg+".toml"))
	if err != nil {
		log.Println("[log] :: install.go | installPkgs() | PkgMakeFromToml failed ➡️ ", err)
		log.Println("[log] ::                            | trying PkgMakeFromViper() ➡️ ", err)
		p, err = PkgMakeFromViper(pkg)
	}

	if err != nil {
		log.Fatalln("[err] :: install.go | installPkgs() or PkgMakeFromViper() | pkg=", pkg, " ➡️ ", err)
	}

	if err = PkgInstall(p, pwd); err != nil {
		log.Fatalln("[err] :: install.go | installPkgs() | PkgInstall() | pkg=", pkg, " ➡️ ", err)
	}

	return nil
}

//----------------------------------------------------------------------------
//                                                          installExtPkgs
//----------------------------------------------------------------------------

// install a package
func installExtPkgs(pwd string) error {
	pkgs := pkgGetExtPkgs()

	if !quiet {
		log.Println("[log] :: install.go | installExtPkgs() | extpkgs ➡️ ", pkgs)
	}

	for _, pkg := range pkgs {

		p, err := PkgMakeFromToml(path.Join(configPath, pkgConfigDir, pkg+".toml"))
		if err != nil {
			p, err = PkgMakeFromViper(pkg)
		}

		if err != nil {
			log.Fatalln("[err] :: install.go | installExtPkgs() | pkg=", pkg, " ➡️ ", err)
		}

		if p.IsExtPkg {
			if err := PkgInstall(p, pwd); err != nil {
				log.Fatalln("[err] :: install.go | PkgInstall() | pkg=", pkg, " ➡️ ", err)
			}
		}

	}

	return nil
}

//----------------------------------------------------------------------------
//                                                                      init
//----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(installCmd)
}
