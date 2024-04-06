package cmd

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

func run_install_command(cargs []string, pkg, step string) {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " installing " + pkg + " cmd: " + cargs[0] + " " + step
	_ = s.Color("red")
	if quiet {
		s.Start() // Start the spinner
		defer s.Stop()
	}

	if !quiet {
		log.Println("[log] :: run_command.go | cmd name ➡️ ", cargs)
	}
	cmd := exec.Command(cargs[0], cargs[1:]...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln("[err] :: run_command.go | cmd.StderrPipe() ➡️ ", err)
	}
	stderr_scanner := bufio.NewScanner(stderr)

	if err := cmd.Start(); err != nil {
		log.Fatalln("[err] :: run_command.go | cmd.Start() ➡️ ", err)
	}

	stderr_scanner.Split(bufio.ScanLines)
	isok := false
	for stderr_scanner.Scan() {
		m := stderr_scanner.Text()
		if !isok {
			isok = strings.Contains(m, "Error") || strings.Contains(m, "error")
		}

		if isok {
			log.Println("[log] :: run_install_command.go: ", m)
		} else if !quiet {
			log.Println("[log] :: run_install_command.go: ", m)
		}
	}
	// stdout_scanner := bufio.NewScanner(stdout)
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatalln("[err] :: run_command.go | cmd.StdoutPipe() ➡️ ", err)
	// }
	// if stderr_scanner.Err() != nil {
	// 	if err := cmd.Process.Kill(); err != nil {
	// 		log.Fatalln("[err] :: run_command.go | cmd.Process.Kill(): ", err)
	// 	}
	//
	// 	if err := cmd.Wait(); err != nil {
	// 		log.Fatalln("[err] :: run_command.go | cmd.Wait(): ", err)
	// 	}
	//
	// 	return
	// }

	// if !quiet {
	// 	stdout_scanner.Split(bufio.ScanLines)
	// 	for stdout_scanner.Scan() {
	// 		log.Println(stdout_scanner.Text())
	// 	}
	//
	// 	if stdout_scanner.Err() != nil {
	// 		if err := cmd.Process.Kill(); err != nil {
	// 			log.Fatalln("[err] :: run_command.go | cmd.Process.Kill(): ", err)
	// 		}
	//
	// 		if err := cmd.Wait(); err != nil {
	// 			log.Fatalln("[err] :: run_command.go | cmd.Wait(): ", err)
	// 		}
	//
	// 		return
	// 	}
	// }

	if err := cmd.Wait(); err != nil {
		log.Fatalln("[err] :: run_command.go | cmd.Wait(): ", err)
	}
}
