package cmd

import (
	"log"
	"os"
)

func change_dir(dir string) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		log.Fatalln("[err] :: make_dir.go | cannot create dir ➡️  ", dir,
			" with permission 0777")
	}

	if err := os.Chdir(dir); err != nil {
		log.Fatalln("[err] :: change_dir.go | os.Chdir() ➡️ ", err)
	}
}
