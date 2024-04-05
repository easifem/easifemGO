package cmd

import (
	"log"
	"os"
	"path"
)

// make director
func make_install_dir(dir string) {
	for _, opt := range []string{"lib", "include", "bin", "share"} {
		astr := path.Join(dir, opt)
		err := os.MkdirAll(astr, 0777)
		if err != nil {
			log.Fatalln("[INTERNAL ERROR] :: make_dir.go | cannot create dir ➡️  ", astr,
				" with permission 0777")
		}
	}
}
