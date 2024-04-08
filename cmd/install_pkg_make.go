package cmd

import (
	"path"
)

func install_pkg_make(pkg, pwd, source_dir, build_dir, install_dir string,
	buildOptions []string,
) {
	change_dir(build_dir)

	cargs := append([]string{
		path.Join(source_dir, "configure"), "--prefix=" + install_dir,
	}, buildOptions...)
	install_run_command(cargs, pkg, "[config]")

	cargs = []string{"make"}
	install_run_command(cargs, pkg, "[build]")

	cargs = []string{"make", "install"}
	install_run_command(cargs, pkg, "[install]")

	change_dir(pwd)
}
