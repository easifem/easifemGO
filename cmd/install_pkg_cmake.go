package cmd

func install_pkg_cmake(pkg, pwd, source_dir, build_dir, install_dir, buildType string,
	buildSharedLibs, buildStaticLibs bool, buildOptions []string,
) {
	cargs := append([]string{
		"cmake",
		"-G", "Ninja",
		"-S", source_dir,
		"-B", build_dir,
		"-D CMAKE_INSTALL_PREFIX:PATH=" + install_dir,
		"-D CMAKE_BUILD_TYPE:STRING=" + _buildType(buildType),
		"-D CMAKE_BUILD_SHARED_LIBS:BOOL=" + _buildSharedLibs(buildSharedLibs),
		"-D CMAKE_BUILD_STATIC_LIBS:BOOL=" + _buildSharedLibs(buildStaticLibs),
	},
		buildOptions...)
	run_install_command(cargs, pkg, "[config]")

	cargs = []string{"cmake", "--build", build_dir}
	run_install_command(cargs, pkg, "[build]")

	cargs = []string{"cmake", "--install", build_dir}
	run_install_command(cargs, pkg, "[install]")
}

func _buildType(a string) string {
	if a == "" {
		return "Release"
	} else {
		return a
	}
}

func _buildSharedLibs(a bool) string {
	if a {
		return "ON"
	} else {
		return "OFF"
	}
}
