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
	cmd := exec.Command(cargs[0], cargs[1:]...)

	output_stdout := make(chan []byte)

	var err error
	go devExecuteCmd(cmd, output_stdout, err)

	var line1, line2, line3, line4, line5 string

	ii := 0

	for data := range output_stdout {
		ii = ii + 1

		switch ii {
		case 1:
			line1 = string(data)
			line2 = line1
			line3 = line1
			line4 = line1
			line5 = line1
		case 2:
			line2 = string(data)
			line3 = line2
			line4 = line2
			line5 = line2
		case 3:
			line3 = string(data)
			line4 = line3
			line5 = line3
		case 4:
			line4 = string(data)
			line5 = line4
		case 5:
			line5 = string(data)
		default:
			line1 = line2
			line2 = line3
			line3 = line4
			line4 = line5
			line5 = string(data)
		}

		found_error := strings.Contains(line5, "Error:")

		if found_error {
			fmt.Println(line1)
			fmt.Println(line2)
			fmt.Println(line3)
			fmt.Println(line4)
			fmt.Println(line5)
			log.Fatalf("[err] :: pkgRunCmd.go | Error found in the output")
		}
	}

	return err
}

//----------------------------------------------------------------------------
//                                                              pkgExecuteCmd
//----------------------------------------------------------------------------

func devExecuteCmd(cmd *exec.Cmd, output_stdout chan []byte, err error) {
	defer close(output_stdout)
	var stdout io.ReadCloser

	if err != nil {
		output_stdout <- []byte(fmt.Sprintf("Error before executing devExecuteCmd(): %v", err))
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
