/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// devCmd represents the dev command
// examples
// easifem dev
var devCmd = &cobra.Command{
	Use:   "dev pkgname [flags]",
	Short: "Develop easifem components",
	Long:  easifem_dev_intro,
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
			log.Fatalln("[err] :: dev.go | os.Getwd() ➡️ ", err)
		}

		pkg := args[0]

		err = devPkgs(pkg, pwd)
		if err != nil {
			log.Fatalln("[err] :: dev.go | devPkgs() ➡ ", err)
		}
	},
}

//----------------------------------------------------------------------------
//                                                            devPkgs
//----------------------------------------------------------------------------

// dev a single package
func devPkgs(pkg, pwd string) error {
	err := PkgDev(easifem_pkgs[pkg], pwd)
	if err != nil {
		return fmt.Errorf("devPkgs() | pkg=%s, err=%w", pkg, err)
	}

	return err
}

//----------------------------------------------------------------------------
//                                                                      init
//----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(devCmd)
}

//----------------------------------------------------------------------------
//                                                                     PkgDev
//----------------------------------------------------------------------------

func PkgDev(pkg *Pkg, pwd string) error {
	if !pkg.IsActive {
		return nil
	}

	for k, v := range pkg.EnvVars {
		os.Setenv(k, v)
	}

	DevCmake(pkg, pwd)

	return nil
}

//----------------------------------------------------------------------------
//                                                                 DevCMake
//----------------------------------------------------------------------------

func DevCmake(pkg *Pkg, pwd string) {
	cargs := []string{
		"cmake",
		"-G", "Ninja",
		"-S", pkg.SourceDir,
		"-B", pkg.BuildDir,
		"-D CMAKE_INSTALL_PREFIX:PATH=" + pkg.InstallDir,
		"-D CMAKE_BUILD_TYPE:STRING=Debug",
		"-D BUILD_SHARED_LIBS:BOOL=ON",
	}

	var err error

	if len(easifem_cache.INSTALL_DIRS) != 0 {
		cargs = append(cargs, "-D CMAKE_PREFIX_PATH:PATH="+strings.Join(easifem_cache.INSTALL_DIRS, ";"))
	}

	cargs = append(cargs, pkg.BuildOptions...)
	err = devRunCmd(cargs)
	if err != nil {
		log.Fatalf("[err] :: pkg.go | PkgCmake() | config step failed %v", err)
	}

	cargs = []string{"cmake", "--build", pkg.BuildDir}
	err = devRunCmd(cargs)
	if err != nil {
		log.Fatalf("[err] :: pkg.go | PkgCmake() | build step failed %v", err)
	}
}

//----------------------------------------------------------------------------
//                                                                pkgRunCmd
//----------------------------------------------------------------------------

func devRunCmd(cargs []string) error {
	if !quiet {
		log.Println("[log] :: pkgRunCmd.go | cmd name ➡️ ", cargs)
	}
	return runRunCmd(cargs, true)
}
