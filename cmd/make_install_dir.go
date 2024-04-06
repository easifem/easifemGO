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
		if err := os.MkdirAll(astr, 0777); err != nil {
			log.Fatalln("[err] :: make_dir.go | cannot create dir ➡️  ", astr,
				" with permission 0777")
		}
	}
}
