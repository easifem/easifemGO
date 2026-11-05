/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package internal

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// fmt.Println(easifem_banner)
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		pkg := args[0]

		err = devPkgs(pkg, pwd)
		if err != nil {
			return err
		}
		return nil
	},
}

//----------------------------------------------------------------------------
//                                                            devPkgs
//----------------------------------------------------------------------------

// dev a single package
func devPkgs(pkg, pwd string) error {
	p, ok := easifem_pkgs[pkg]
	if !ok {
		return fmt.Errorf("unknown package %q", pkg)
	}
	started := time.Now()
	err := PkgDev(p, pwd)
	if err != nil {
		return fmt.Errorf("devPkgs() | pkg=%s, err=%w", pkg, err)
	}

	if p.IsActive && !quiet {
		fmt.Fprintf(os.Stderr, "\n  ✓ %s built successfully · %s\n", pkg, time.Since(started).Round(100*time.Millisecond))
	}
	return nil
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

	return DevCmake(pkg, pwd)
}

//----------------------------------------------------------------------------
//                                                                 DevCMake
//----------------------------------------------------------------------------

func DevCmake(pkg *Pkg, pwd string) (err error) {
	output := newLibraryOutput(pkg.Name)
	defer func() { output.finish(err) }()
	cargs := []string{
		"cmake",
		"-G", "Ninja",
		"-S", pkg.SourceDir,
		"-B", pkg.BuildDir,
		"-D CMAKE_INSTALL_PREFIX:PATH=" + pkg.InstallDir,
		"-D CMAKE_BUILD_TYPE:STRING=Debug",
		"-D BUILD_SHARED_LIBS:BOOL=ON",
	}

	if len(easifem_cache.INSTALL_DIRS) != 0 {
		cargs = append(cargs, "-D CMAKE_PREFIX_PATH:PATH="+strings.Join(easifem_cache.INSTALL_DIRS, ";"))
	}

	cargs = append(cargs, pkg.BuildOptions...)
	err = output.run(cargs, "Configure", pwd)
	if err != nil {
		return err
	}

	cargs = []string{"cmake", "--build", pkg.BuildDir}
	err = output.run(cargs, "Build", pwd)
	if err != nil {
		return err
	}
	return nil
}

//----------------------------------------------------------------------------
//                                                                pkgRunCmd
//----------------------------------------------------------------------------

func devRunCmd(cargs []string) error {
	return runBuildStage(cargs, "easifem", "Build", "")
}
