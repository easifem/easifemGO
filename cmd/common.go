package cmd

const easifem_banner string = `
    ███████╗ █████╗ ███████╗██╗███████╗███████╗███╗   ███╗
    ██╔════╝██╔══██╗██╔════╝██║██╔════╝██╔════╝████╗ ████║
    █████╗  ███████║███████╗██║█████╗  █████╗  ██╔████╔██║
    ██╔══╝  ██╔══██║╚════██║██║██╔══╝  ██╔══╝  ██║╚██╔╝██║
    ███████╗██║  ██║███████║██║██║     ███████╗██║ ╚═╝ ██║
    ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝╚═╝     ╚══════╝╚═╝     ╚═╝
    Expandable And Scalable Infrastructure for Finite Element Methods
    (c) Vikas Sharma, Ph.D., vickysharma0812@gmail.com
    https://www.easifem.com
    version 24.3.0 --HEAD
    `

const easifem_intro string = `
Expandable And Scalable Infrastructure for Finite Element Methods
===============================================================
  easifem is a CLI (Command Line Interface) for working with
  EASIFEM platform. It is written in Go language. It contains
  many subcommands which will help user to integrate EASIFEM
  with their projects.
  The user can perform following actions:

  - Set environment variable for EASIFEM platform
  - Install/Uninstall/Reinstall components of EASIFEM
  - Run application which uses EASIFEM

  For more information visit:
  website: https://www.easifem.com
  (c) Vikas Sharma, vickysharma0812@gmail.com
`

const easifem_setenv_intro = `
The [setenv] subcommand sets the environment variable for easifem on your system.

While setting the environment you can provide following details.
A) EASIFEM_INSTALL_DIR
B) EASIFEM_BUILD_DIR
C) EASIFEM_SOURCE_DIR

EASIFEM_INSTALL_DIR: denotes the root location where EASIFEM will
be installed. It is specified by --install=value

Following are the good choices for --install variable:
1) ${HOME}
2) ${HOME}/.local
3) /opt
The default value is ${HOME}.

EASIFEM_SOURCE_DIR: specifies the location where the source code of
the components of EASIFEM will be stored.
It is specified by --source=value

Following are the good choices for --source variable:
1) ${HOME}/code/
The default value is ${HOME}/code.
EASIFEM_BUILD_DIR: specifies the location where the components of
EASIFEM will be build. It is specified by --build=value

Following are the good choices for --root variable:
1) ${HOME}/temp
The default value is ${HOME}/temp.

Example:

easifem setenv --install ${HOME} --build ${HOME}/temp --source ${HOME}/code
easifem setenv -r ${HOME} -b ${HOME}/temp -s ${HOME}/code
`

const easifem_install_intro = `
The [install] subcommand helps you installing the components of
EASIFEM, such as extpkgs, base, classes, materials, kernels, etc.

In order to install a component you should specify following environment variables:
A) EASIFEM_ROOT_DIR: the place where easifem is installed.
B) EASIFEM_BUILD_DIR: the place where easifem is build.
C) EASIFEM_SOURCE_DIR: the place where the source of easifem will be stored.

You can specify them by using

easifem setenv --build= --install= --source=

For more see,
easifem setenv --help

Use:

easifem install option

Option can be:

all: Install everything
extpkgs: install external packages
openblas : install OpenBlas
lapack95: install LAPACK95
sparsekit: install Sparsekit
fftw: install FFTW
superlu: install SuperLU
arpack: install ARPACK
lis: install LIS
base: install easifemBase
classes: install easifemClasses
materials: install easifemMaterials
kernels: install easifemKernels
`

const easifem_dev_intro = `
The [dev] subcommand helps you in developing easifem-components,
such as base, classes, materials, kernels, etc.

  The library will not install the package. It will just build it, 
  and show warnings and errors.
  It will always build in the debug mode.
  We never download the project.

HOW TO USE:
  easifem dev base
  easifem dev classes
`

const easifem_lint_intro = `
The [lint] subcommand helps you linting the easifem files while building the easifem.
This command is only for developers. The end users should not use this command.

easifem lint filename [projectname]

projectname can be "base", "classes", or "all"
Default is "all"
`

const easifem_run_intro = `
The [run] subcommand helps you build and run the applications built using easifem.

The [run] subcommand compiles and links an application, which is build
by using easifem components, such as, base, classes, etc.
This program can also PARSE a fortran code kept in the code-fences of 
markdown file with extension [.md]
The example of code fence which can be used in the markdown file is given below.
Any thing outside fortran code fence will be ignored by the parser.
Also you can use as many fences as you want, good for Documentation.

To run this file  you can use:

easifem run filename [projectname]

projectname can be "base", "classes", or "all"

Default is "all"
`

const easifem_clean_intro = "clean the build directory,  keep source code."

const (
	// easifem_version_major  int    = 23
	// easifem_version_minor  int    = 10
	// easifem_version_patch  int    = 5
	// easifem_version_string string = "23.10.05"

	easifem_config_dir       = "easifem"
	easifem_config_name      = "easifem"
	easifem_config_type      = "toml"
	easifem_install_dir      = "$HOME/.easifem/install"
	easifem_build_dir        = "$HOME/.easifem/build"
	easifem_source_dir       = "$HOME/.easifem/src"
	easifem_lint_dir         = "$HOME/.easifem/lint"
	easifem_default_env_name = "env"
	easifem_build_type       = "Release" // default value of buil;d type
	easifem_pkg_config_dir   = "plugins"
	easifem_build_system     = "cmake"
	easifem_cache_file       = "cache.toml"
	easifem_shell_var_file   = "easifemvar"
)

var (
	quiet                    bool
	noDownload               bool
	sourceDir                string
	buildDir                 string
	installDir               string
	lintDir                  string
	configPath               string // config file name with extension
	configFile               string // config file name with extension
	easifem_current_env_name = easifem_default_env_name
)

type Cache struct {
	LD_LIBRARY_PATH []string // LD_LIBRARY_PATH
	INSTALL_DIRS    []string // INSTALL DIRS
}

var easifem_cache = Cache{}

//----------------------------------------------------------------------------
//                                                                       Pkg
//----------------------------------------------------------------------------

type Pkg struct {
	BuildDir        string
	BuildOptions    []string
	BuildSharedLibs bool
	BuildStaticLibs bool
	BuildSystem     string
	BuildType       string
	CmakePrefixPath []string
	Dependencies    []string
	EnvVars         map[string]string
	Git             string
	InstallDir      string
	IsActive        bool
	IsExtPkg        bool
	LdLibraryPath   []string
	License         string
	Name            string
	RunTest         bool
	SourceDir       string
	TargetName      string
	Url             string
}

var easifem_pkgs = make(map[string]*Pkg)

//----------------------------------------------------------------------------
//                                                                 Linter
//----------------------------------------------------------------------------

type Linter struct {
	IncludePath  []string
	Flags        []string // compiler flags passed during build stage
	Compiler     string
	CompilerPath string
	LintDir      string // Where to put lint files (.smod, .mod, .o)
}

//----------------------------------------------------------------------------
//                                                                    Runner
//----------------------------------------------------------------------------

type Runner struct {
	BuildDir             string
	BuildType            string // Release or Debug
	CMakeMinimumVersion  string
	Compiler             string
	CompilerPath         string
	ExtraCMakePrefixPath []string // These are extra CMakePrefixPaths
	ExtraIncludePath     []string // These are additional paths passed to the compiler while building
	ExtraLibs            []string // These libraries should be either the pull path or they should be installed where ld can find them
	FileName             string   // name of main.F90
	Flags                []string
	IncludePath          []string
	Language             string
	LibraryPath          []string
	ProjectName          string
	SourceDir            string
	TargetLibs           []string // These libraries are build using Cmake and Cmake can find theme
	TargetName           string   // name of executable
}

const (
	gfortranArgs        = `"-ffree-form" "-ffree-line-length-none" "-std=f2008" "-fimplicit-none"`
	gfortranReleaseArgs = `"-O3"`
	gfortranDebugArgs   = ` "-fbounds-check" "-g" "-fbacktrace" "-Wextra" "-Wall" "-fprofile-arcs" "-ftest-coverage" "-Wimplicit-interface" `

	intelArgs        = `"-r8" "-W1"`
	intelReleaseArgs = `"-O3"`
	intelDebugArgs   = `"-O0" "-traceback" "-g" "-debug all" "-check all" "-ftrapuv" "-warn" "nointerfaces"`

	xlfArgs        = `"-q64" "-qrealsize=8" "-qsuffix=f=f90:cpp=f90"`
	xlfReleaseArgs = `"-O3" "-qstrict"`
	xlfDebugArgs   = `"-O0" "-g" "-qfullpath" "-qkeepparm"`
)
