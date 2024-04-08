package cmd

type version struct {
	Major   int
	Minor   int
	Patch   int
	VerName string
}

type Project struct {
	MinCMakeVersion string
	Name            string
	Homepage        string
	TargetName      string
	NameSpace       string
	CompilerDef     []string
	Lang            []string
	Version         version
	SharedLibs      bool
	StaticLibs      bool
	OpenBlas        bool
	OpenMP          bool
	SuperLU         bool
	Lis             bool
	Hdf5            bool
	Plplot          bool
	Gmsh            bool
	Fftw            bool
	GtkFortran      bool
	Lua             bool
	IntSize         int // integer size in bytes
	RealSize        int // real size in byte
}

func DumpCmake(p Project) string {
	ans := `
cmake_minimum_required(VERSION 3.20.0 FATAL_ERROR)
  `

	return ans
}
