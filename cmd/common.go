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

const easifem_clean_intro = "clean the build directory,  keep source code."

const (
	// easifem_version_major  int    = 23
	// easifem_version_minor  int    = 10
	// easifem_version_patch  int    = 5
	// easifem_version_string string = "23.10.05"

	easifem_config_dir  = "easifem"
	easifem_config_name = "easifem"
	easifem_config_type = "toml"

	easifem_install_dir      = "$HOME/.easifem/install"
	easifem_build_dir        = "$HOME/.easifem/build"
	easifem_source_dir       = "$HOME/.easifem/src"
	easifem_default_env_name = "env"

	easifem_build_type = "Release" // default value of buil;d type
)

var (
	quiet bool
	// debugMode  bool
	sourceDir                string
	buildDir                 string
	installDir               string
	configPath               string             // config file name with extension
	configFile               string             // config file name with extension
	buildType                string = "Release" // build type "Release", "Debug"
	buildSharedLibs                 = true
	buildStaticLibs                 = false
	easifem_current_env_name        = easifem_default_env_name
)
