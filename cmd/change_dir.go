package cmd

import (
	"log"
	"os"
)

func change_dir(dir string) {
	if err := os.Chdir(dir); err != nil {
		log.Fatalln("[INTERNAL ERROR] :: change_dir.go | os.Chdir() ➡️ ", err)
	}
}
