/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package internal

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
// examples
// easifem lint filename projectname
// project name is optional ["base", "classes"], default is "classes"

var lintCmd = &cobra.Command{
	Use:   "lint filename [flags]",
	Short: "Linting easifem project while dev ops",
	Long:  easifem_lint_intro,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: lint.go | os.Getwd() ➡️ ", err)
		}

		filename := args[0]
		projectname := "all"

		if len(args) >= 2 {
			projectname = args[1]
		}

		err = lint(filename, projectname, pwd)
		if err != nil {
			log.Fatalln("[err] :: lint.go | lint() ➡ ", err)
		}
	},
}

//----------------------------------------------------------------------------
//                                                            lintPkgs
//----------------------------------------------------------------------------

// Built-in diagnostic flags; Flags from linter.toml are appended after them.
var lintDefaultFlags = []string{
	"-ffree-form",
	"-ffree-line-length-none",
	"-std=f2018",
	"-fimplicit-none",
	"-Waliasing",
	"-Wall",
	"-Wampersand",
	"-Warray-bounds",
	"-Wc-binding-type",
	"-Wcharacter-truncation",
	"-Wconversion",
	"-Wdo-subscript",
	"-Wfunction-elimination",
	"-Wimplicit-interface",
	"-Wimplicit-procedure",
	"-Wintrinsic-shadow",
	"-Wuse-without-only",
	"-Wintrinsics-std",
	"-Wline-truncation",
	"-Wno-align-commons",
	"-Wno-overwrite-recursive",
	"-Wno-tabs",
	"-Wreal-q-constant",
	"-Wsurprising",
	"-Wunderflow",
	"-Wunused-parameter",
	"-Wrealloc-lhs",
	"-Wrealloc-lhs-all",
	"-Wtarget-lifetime",
	"-pedantic",
	"-pedantic-errors",
}

// lint compiles a single file with strict diagnostics.
func lint(filename, projectname, pwd string) error {
	linter := Linter{
		Compiler:     "gfortran",
		CompilerPath: "",
		LintDir:      lintDir,
	}

	tomlFile := path.Join(pwd, "linter.toml")
	if _, err := os.Stat(tomlFile); err == nil {
		if _, err := toml.DecodeFile(tomlFile, &linter); err != nil {
			return fmt.Errorf("read %s: %w", tomlFile, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	cargs, err := lintArgs(linter, filename, pwd, easifem_cache.INSTALL_DIRS)
	if err != nil {
		return err
	}

	if err := makePackageDirs(os.ExpandEnv(linter.LintDir)); err != nil {
		return err
	}

	if err := lintRunCmd(cargs, filename, "[building]"); err != nil {
		return fmt.Errorf("lint %s: %w", filename, err)
	}
	return nil
}

// lintArgs builds the compiler command line: compiler, include paths from
// linter.toml, the include directory of each installed prefix, built-in
// flags, Flags from linter.toml, then module/object output options.
func lintArgs(linter Linter, filename, pwd string, installDirs []string) ([]string, error) {
	compiler, err := fortranCompiler(linter.Compiler, linter.CompilerPath, pwd)
	if err != nil {
		return nil, err
	}
	if compiler == "" {
		compiler = "gfortran"
	}

	cargs := []string{compiler}
	for _, incpath := range linter.IncludePath {
		cargs = append(cargs, "-I", os.ExpandEnv(incpath))
	}
	for _, installDir := range installDirs {
		cargs = append(cargs, "-I", path.Join(installDir, "include"))
	}
	cargs = append(cargs, lintDefaultFlags...)
	for _, flag := range linter.Flags {
		cargs = append(cargs, os.ExpandEnv(flag))
	}

	lintDir := os.ExpandEnv(linter.LintDir)
	cargs = append(cargs,
		"-J", path.Join(lintDir, "include"),
		"-c", filename,
		"-o", path.Join(lintDir, "lib", filepath.Base(filename)+".o"),
	)
	return cargs, nil
}

//----------------------------------------------------------------------------
//                                                                      init
//----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(lintCmd)
}

//----------------------------------------------------------------------------
//                                                                lintRunCmd
//----------------------------------------------------------------------------

func lintRunCmd(cargs []string, pkg, step string) error {
	if !quiet {
		log.Println("[log] :: pkgRunCmd.go | cmd name ➡️ ", cargs)
	}
	return runRunCmd(cargs, false)
}
