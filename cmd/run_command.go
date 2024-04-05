package cmd

import (
	"log"
	"os/exec"
)

func run_command(cargs []string) {
	if !quiet {
		log.Println("[log] :: run_command.go | cmd name ➡️ ", cargs)
	}
	cmd := exec.Command(cargs[0], cargs[1:]...)
	err := cmd.Run()
	if err != nil {
		stdoutStderr, _ := cmd.CombinedOutput()
		log.Printf("[log] :: run_command.go | cmd output ➡️ %s\n", stdoutStderr)
		log.Fatalln("[INTERNAL ERROR] :: run_command.go() | cmd.Run() ➡️ ", err)
	}
	if !quiet {
		out, _ := cmd.Output()
		log.Printf("[log] :: run_command.go | cmd output ➡️ %s\n", out)
	}
}
