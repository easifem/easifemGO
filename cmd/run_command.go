package cmd

import (
	"bufio"
	"log"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
)

func run_install_command(cargs []string) {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " command is running"
	_ = s.Color("red")

	if !quiet {
		log.Println("[log] :: run_command.go | cmd name ➡️ ", cargs)
	}
	cmd := exec.Command(cargs[0], cargs[1:]...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.StderrPipe() ➡️ ", err)
	}
	stderr_scanner := bufio.NewScanner(stderr)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.StdoutPipe() ➡️ ", err)
	}
	stdout_scanner := bufio.NewScanner(stdout)

	if quiet {
		s.Start() // Start the spinner
	}

	if err := cmd.Start(); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Start() ➡️ ", err)
		if quiet {
			s.Stop() // Start the spinner
		}
	}

	for stderr_scanner.Scan() {
		// Do something with the line here.
		// ProcessLine(scanner.Text())
	}

	if stderr_scanner.Err() != nil {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Process.Kill(): ", err)
		}

		if err := cmd.Wait(); err != nil {
			log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Wait(): ", err)
		}

		if quiet {
			s.Stop()
		}
		return
	}

	if !quiet {
		for stdout_scanner.Scan() {
			// Do something with the line here.
			// ProcessLine(scanner.Text())
		}

		if stdout_scanner.Err() != nil {
			if err := cmd.Process.Kill(); err != nil {
				log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Process.Kill(): ", err)
			}

			if err := cmd.Wait(); err != nil {
				log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Wait(): ", err)
			}

			if quiet {
				s.Stop()
			}
			return
		}
	}

	if err := cmd.Wait(); err != nil {
		if quiet {
			s.Stop()
		}
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Wait(): ", err)
	}

	if quiet {
		s.Stop()
	}
}
