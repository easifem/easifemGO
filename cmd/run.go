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

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
// examples
// easifem run filename projectname
// project name is optional ["base", "classes"], default is "classes"

var runCmd = &cobra.Command{
	Use:   "run filename [flags]",
	Short: "build and run executable file.",
	Long:  easifem_run_intro,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("[err] :: run.go | os.Getwd() ➡️ ", err)
		}

		files := ""
		for _, file := range args {
			files = files + " " + file
		}

		err = run(files, pwd)
		if err != nil {
			log.Fatalln("[err] :: run.go | run() ➡ ", err)
		}
	},
}

//----------------------------------------------------------------------------
//                                                            runPkgs
//----------------------------------------------------------------------------

// run a single package
func run(filename, pwd string) error {
	var err error

	runner := Runner{
		BuildDir:            "build",
		BuildType:           "Release",
		CMakeMinimumVersion: "3.20.0",
		Compiler:            "gfortran",
		FileName:            filename,
		Language:            "Fortran",
		ProjectName:         "easifemApp",
		SourceDir:           pwd,
		TargetName:          "main.out",
		IsExecute:           noRun,
		CacheClean:          cacheClean,
		ReBuild:             reBuild,
	}

	tomlFile := path.Join(pwd, "runner.toml")

	if _, err = os.Stat(tomlFile); !os.IsNotExist(err) {
		_, _ = toml.DecodeFile(tomlFile, &runner)
	}

	runner.BuildDir = os.ExpandEnv(runner.BuildDir)
	runner.SourceDir = os.ExpandEnv(runner.SourceDir)

	if len(runner.TargetLibs) == 0 {
		runner.TargetLibs = append(runner.TargetLibs, "easifemBase")
	}

	cmake := makeCmakeFile(runner)
	err = writeCmakeFile(cmake, pwd)
	if err != nil {
		return fmt.Errorf("run() | error = %w", err)
	}

	cargs := []string{
		"cmake",
		"-G", "Ninja",
		"-S", runner.SourceDir,
		"-B", runner.BuildDir,
		"-D CMAKE_BUILD_TYPE:STRING=" + runner.BuildType,
	}

	if runner.CacheClean {
		cargs = append(cargs, "--fresh")
	}

	log.Println("running cmd: ", cargs)
	err = runRunCmd(cargs, true)
	if err != nil {
		return fmt.Errorf("run() | error = %w", err)
	}

	cargs = []string{"cmake", "--build", runner.BuildDir}

	if runner.ReBuild {
		cargs = append(cargs, "--clean-first")
	}

	log.Println("running cmd: ", cargs)
	err = runRunCmd(cargs, true)
	if err != nil {
		return fmt.Errorf("run() | error = %w", err)
	}

	cargs = []string{path.Join(runner.BuildDir, runner.TargetName)}
	if runner.IsExecute {
		log.Println("executable file path: ", cargs)
		return err
	}

	log.Println("running cmd: ", cargs)
	err = runRunCmd(cargs, false)
	if err != nil {
		return fmt.Errorf("run() | error = %w", err)
	}

	return err
}

//----------------------------------------------------------------------------
//                                                                      init
//----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.PersistentFlags().BoolVar(&noRun, "no-run", false,
		"Only create the binary file and do not run it.")
	rootCmd.PersistentFlags().BoolVar(&cacheClean, "cache-clean", false,
		"Clean all cache files and recreate them from scratch")
	rootCmd.PersistentFlags().BoolVar(&reBuild, "rebuild", false,
		"Clean the build directory first and then build")
}

//----------------------------------------------------------------------------
//                                                                runRunCmd
//----------------------------------------------------------------------------

func makeCmakeFile(runner Runner) []string {
	cmake_prefix_path := ""
	for _, path := range easifem_cache.INSTALL_DIRS {
		cmake_prefix_path = cmake_prefix_path + fmt.Sprintf(" %q ", path)
	}

	for _, path := range runner.ExtraCMakePrefixPath {
		cmake_prefix_path = cmake_prefix_path + fmt.Sprintf(" %q ", path)
	}

	cmake := []string{
		fmt.Sprintf("cmake_minimum_required(VERSION %s FATAL_ERROR)\n", runner.CMakeMinimumVersion),
		fmt.Sprintf("project(%q)\n", runner.ProjectName),
		fmt.Sprintf("enable_language(%s)\n", runner.Language),
		`
    if (NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Debug CACHE STRING "Build type" FORCE)
    endif()
    `,
		fmt.Sprintf("if(${CMAKE_Fortran_COMPILER_ID} STREQUAL %q OR Fortran_COMPILER_NAME MATCHES %q)\n", "GNU", "gfortran*"),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS %s)\n", gfortranArgs),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS_RELEASE %s)\n", gfortranReleaseArgs),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS_DEBUG %s)\n", gfortranDebugArgs),
		"endif()\n",
		fmt.Sprintf("if(${CMAKE_Fortran_COMPILER_ID} STREQUAL %q OR Fortran_COMPILER_NAME MATCHES %q)\n", "Intel", "ifort*"),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS %s)\n", intelArgs),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS_RELEASE %s)\n", intelReleaseArgs),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS_DEBUG %s)\n", intelDebugArgs),
		"endif()\n",
		fmt.Sprintf("if(${CMAKE_Fortran_COMPILER_ID} STREQUAL %q OR Fortran_COMPILER_NAME MATCHES %q)\n", "XL", "xlf*"),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS %s)\n", xlfArgs),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS_RELEASE %s)\n", xlfReleaseArgs),
		fmt.Sprintf("list(APPEND FORTRAN_FLAGS_DEBUG %s)\n", xlfDebugArgs),
		"endif()\n",
		fmt.Sprintf("list(APPEND CMAKE_PREFIX_PATH %s)\n", cmake_prefix_path),
	}

	for _, lib := range runner.TargetLibs {
		cmake = append(cmake, fmt.Sprintf("find_package(%q)\n", lib))
	}

	for _, file := range runner.ExtraFileNames {
		runner.FileName = fmt.Sprintf("%s %s", runner.FileName, file)
	}

	cmake = append(cmake, fmt.Sprintf("add_executable(%s %s)\n", runner.TargetName, runner.FileName))

	target_link_libs := ""
	for _, lib := range runner.TargetLibs {
		target_link_libs = target_link_libs + fmt.Sprintf(" %s::%s ", lib, lib)
	}

	for _, lib := range runner.ExtraLibs {
		target_link_libs = target_link_libs + fmt.Sprintf(" %q ", lib)
	}

	cmake = append(cmake, fmt.Sprintf("target_link_libraries(%s PRIVATE %s)\n", runner.TargetName, target_link_libs))

	return cmake
}

//----------------------------------------------------------------------------
//
//----------------------------------------------------------------------------

// Read the cache
func writeCmakeFile(cmake []string, pwd string) error {
	afile := path.Join(pwd, "CMakeLists.txt")

	f, err := os.Create(afile)
	if err != nil {
		return fmt.Errorf("writeCmakeFile() | error = %w", err)
	}

	defer f.Close()

	for _, line := range cmake {
		_, err = f.WriteString(line)
	}

	return err
}

//----------------------------------------------------------------------------
//                                                                lintRunCmd
//----------------------------------------------------------------------------

func runRunCmd(cargs []string, qmode bool) error {
	cmd := exec.Command(cargs[0], cargs[1:]...)

	output_stdout := make(chan []byte)

	var err error
	go runExecuteCmd(cmd, output_stdout, err)

	if err != nil {
		return err
	}

	if qmode {
		for data := range output_stdout {
			log.Println(string(data))
		}
	} else {
		for data := range output_stdout {
			fmt.Println(string(data))
		}
	}

	return err
}

//----------------------------------------------------------------------------
//                                                              pkgExecuteCmd
//----------------------------------------------------------------------------

func runExecuteCmd(cmd *exec.Cmd, output_stdout chan []byte, err error) {
	defer close(output_stdout)
	var stdout io.ReadCloser

	if err != nil {
		output_stdout <- []byte(fmt.Sprintf("Error executing: %v", err))
		return
	}

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		output_stdout <- []byte(fmt.Sprintf("Error getting stdout pipe: %v", err))
		return
	}

	cmd.Stderr = cmd.Stdout
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
