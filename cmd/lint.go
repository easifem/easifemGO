/*
Copyright © 2024 Vikas Sharma vickysharma0812@gmail.com
*/

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

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

// lint a single package
func lint(filename, projectname, pwd string) error {
	var err error

	linter := Linter{
		Compiler:     "gfortran",
		CompilerPath: "",
		LintDir:      lintDir,
	}

	tomlFile := path.Join(pwd, "linter.toml")

	if _, err = os.Stat(tomlFile); !os.IsNotExist(err) {
		_, _ = toml.DecodeFile(tomlFile, &linter)
	}

	tsize := 0
	tsize = len(linter.IncludePath)
	temp_inc_path_list := make([]string, 2*tsize)

	for ii, incpath := range linter.IncludePath {
		temp_inc_path_list[2*ii] = "-I"
		temp_inc_path_list[2*ii+1] = os.ExpandEnv(incpath)
	}

	// linter.CompilerPath = ""
	tsize = tsize + len(easifem_cache.INSTALL_DIRS)
	linter.IncludePath = make([]string, 2*tsize)

	for ii, v := range temp_inc_path_list {
		linter.IncludePath[ii] = v
	}

	tsize = len(temp_inc_path_list)
	for ii, installDir := range easifem_cache.INSTALL_DIRS {
		linter.IncludePath[2*ii+tsize] = "-I"
		linter.IncludePath[2*ii+tsize+1] = path.Join(installDir, "include")
	}

	linter.Flags = []string{
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

	cargs := []string{
		linter.Compiler,
	}

	cargs = append(cargs, linter.IncludePath...)
	cargs = append(cargs, linter.Flags...)

	linter.LintDir = os.ExpandEnv(linter.LintDir)
	pkgMakeDir(linter.LintDir)

	// clean_filename := path.Join(pwd, filename)
	clean_filename := strings.ReplaceAll(filename, " ", "\\ ")
	obj_output := path.Join(linter.LintDir, "lib", filepath.Base(filename)+".o")

	cargs = append(cargs,
		"-J",
		path.Join(linter.LintDir, "include"),
		"-c",
		clean_filename,
		"-o",
		obj_output,
	)

	err = lintRunCmd(cargs, filename, "[building]")
	if err != nil {
		log.Fatalf("[err] :: lint.go | lint() | building failed %v", err)
	}

	return err
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
	cmd := exec.Command(cargs[0], cargs[1:]...)

	output_stdout := make(chan []byte)

	var err error
	go lintExecuteCmd(cmd, output_stdout, err)

	for data := range output_stdout {
		fmt.Println(string(data))
	}

	return err
}

//----------------------------------------------------------------------------
//                                                              pkgExecuteCmd
//----------------------------------------------------------------------------

func lintExecuteCmd(cmd *exec.Cmd, output_stdout chan []byte, err error) {
	defer close(output_stdout)
	var stdout io.ReadCloser

	stdout, err = cmd.StderrPipe()
	if err != nil {
		log.Println(err)
		output_stdout <- []byte(fmt.Sprintf("Error getting stdout pipe: %v", err))
		return
	}

	// cmd.Stderr = cmd.Stdout
	stdout_scanner := bufio.NewScanner(stdout)

	done := make(chan struct{})

	err = cmd.Start()
	if err != nil {
		output_stdout <- []byte(fmt.Sprintf("Error executing: %v", err))
		return
	}

	go func() {
		for stdout_scanner.Scan() {
			output_stdout <- stdout_scanner.Bytes()
		}
		done <- struct{}{}
	}()

	<-done

	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
		output_stdout <- []byte(fmt.Sprintf("Error waiting for the script to complete: %v", err))
	}
}
