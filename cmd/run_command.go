package cmd

import (
	"io"
	"log"
	"os/exec"
)

func run_command(cargs []string) {
	if !quiet {
		log.Println("[log] :: run_command.go | cmd name ➡️ ", cargs)
	}
	cmd := exec.Command(cargs[0], cargs[1:]...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.StderrPipe() ➡️ ", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Start() ➡️ ", err)
	}

	slurp, _ := io.ReadAll(stderr)
	if len(slurp) != 0 {
		log.Printf("[log] :: run_command.go | stderr %s\n", slurp)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: run_command.go | cmd.Wait() ➡️ ", err)
	}

	if !quiet {
		out, _ := cmd.Output()
		if len(out) != 0 {
			log.Printf("[log] :: run_command.go | cmd output ➡️ %s\n", out)
		}
	}
}
