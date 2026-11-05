package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"reflect"
	"sort"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/spf13/viper"
)

//----------------------------------------------------------------------------
//                                                                    PkgMake
//----------------------------------------------------------------------------

func PkgMake(pkg *Pkg, pwd string) (err error) {
	output := newLibraryOutput(pkg.Name)
	defer func() { output.finish(err) }()
	return pkgMakeWithOutput(pkg, pwd, output)
}

func pkgMakeWithOutput(pkg *Pkg, pwd string, output *libraryOutput) error {
	if err := os.MkdirAll(pkg.BuildDir, 0777); err != nil {
		return err
	}

	cargs := []string{
		path.Join(pkg.SourceDir, "configure"),
		"--prefix=" + pkg.InstallDir,
	}

	var err error

	cargs = append(cargs, pkg.BuildOptions...)
	err = output.run(cargs, "Configure", pkg.BuildDir)
	if err != nil {
		return err
	}

	cargs = []string{"make"}
	err = output.run(cargs, "Build", pkg.BuildDir)
	if err != nil {
		return err
	}

	cargs = []string{"make", "install"}
	err = output.run(cargs, "Install", pkg.BuildDir)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
//                                                              PkgCMake
//----------------------------------------------------------------------------

func PkgCmake(pkg *Pkg, pwd string) (err error) {
	output := newLibraryOutput(pkg.Name)
	defer func() { output.finish(err) }()
	return pkgCmakeWithOutput(pkg, pwd, output)
}

func pkgCmakeWithOutput(pkg *Pkg, pwd string, output *libraryOutput) error {
	cargs := []string{
		"cmake",
		"-G", "Ninja",
		"-S", pkg.SourceDir,
		"-B", pkg.BuildDir,
		"-D CMAKE_INSTALL_PREFIX:PATH=" + pkg.InstallDir,
		"-D CMAKE_BUILD_TYPE:STRING=" + pkg.BuildType,
		"-D BUILD_SHARED_LIBS:BOOL=" + cmakeOnOff(pkg.BuildSharedLibs),
		"-D BUILD_STATIC_LIBS:BOOL=" + cmakeOnOff(pkg.BuildStaticLibs),
	}

	var err error

	if len(easifem_cache.INSTALL_DIRS) != 0 {
		cargs = append(cargs, "-D CMAKE_PREFIX_PATH:PATH="+strings.Join(easifem_cache.INSTALL_DIRS, ";"))
	}

	cargs = append(cargs, pkg.BuildOptions...)
	err = output.run(cargs, "Configure", pwd)
	if err != nil {
		return err
	}

	cargs = []string{"cmake", "--build", pkg.BuildDir}
	err = output.run(cargs, "Build", pwd)
	if err != nil {
		return err
	}

	cargs = []string{"cmake", "--install", pkg.BuildDir}
	err = output.run(cargs, "Install", pwd)
	if err != nil {
		return err
	}
	return nil
}

//----------------------------------------------------------------------------
//                                                                pkgMason
//----------------------------------------------------------------------------

// Install package usign mason build system
func PkgMason(pkg *Pkg, pwd string) {
}

//----------------------------------------------------------------------------
//                                                                 PkgInstall
//----------------------------------------------------------------------------

func PkgInstall(pkg *Pkg, pwd string) (err error) {
	if !pkg.IsActive {
		return nil
	}

	output := newLibraryOutput(pkg.Name)
	defer func() { output.finish(err) }()
	if !noDownload {
		if err := pkgDownloadPkgWithOutput(pkg.Url, pkg.SourceDir, pwd, output); err != nil {
			return err
		}
	}

	if err := makePackageDirs(pkg.InstallDir); err != nil {
		return err
	}

	for k, v := range pkg.EnvVars {
		os.Setenv(k, v)
	}

	switch pkg.BuildSystem {
	case "make":
		if err := pkgMakeWithOutput(pkg, pwd, output); err != nil {
			return err
		}
	case "cmake":
		if err := pkgCmakeWithOutput(pkg, pwd, output); err != nil {
			return err
		}
	case "mason":
		PkgMason(pkg, pwd)
	default:
	}

	return nil
}

//----------------------------------------------------------------------------
//                                                            PkgMakeFromToml
//----------------------------------------------------------------------------

// Make pkg from toml file
func PkgMakeFromToml(pkg string) (*Pkg, error) {
	var p *Pkg
	var err error
	tomlFile := path.Join(configPath, easifem_pkg_config_dir, pkg+".toml")

	if verbose && !quiet {
		log.Printf("[log] :: pkg.go | PkgMakeFromToml() pkg=%s, file=%s",
			pkg, tomlFile)
	}

	p = &Pkg{IsActive: true, IsExtPkg: false}
	_, err = toml.DecodeFile(tomlFile, p)
	if err != nil {
		return p, err
	}

	if err = PkgCheckAndFix(p); err != nil {
		return p, err
	}

	if err != nil {
		p, err = PkgMakeFromViper(pkg)
	}

	p.BuildDir = os.ExpandEnv(p.BuildDir)
	p.InstallDir = os.ExpandEnv(p.InstallDir)
	p.SourceDir = os.ExpandEnv(p.SourceDir)

	return p, err
}

//----------------------------------------------------------------------------
//                                                                PkgLogPrint
//----------------------------------------------------------------------------

// This function prints the pkg by using log.Println
func PkgLogPrint(p *Pkg) {
	log.Println("[log] :: pkg.go | PkgLogPrint()")
	indent := "[log] :: pkg.go | PkgLogPrint() pkg." + p.Name + "."
	typ, val := reflect.TypeOf(p), reflect.ValueOf(p)
	for i := 0; i < typ.NumField(); i++ {
		indent := indent
		log.Printf("%s%-11s → %v\n", indent, typ.Field(i).Name, val.Field(i).Interface())
	}
}

//----------------------------------------------------------------------------
//                                                          PkgCheckAndFix
//----------------------------------------------------------------------------

// Check and fix the pkg
func PkgCheckAndFix(pkg *Pkg) error {
	var err error

	if !pkg.IsActive {
		return nil
	}

	if pkg.Name == "" {
		return errors.New("pkg.Name is empty")
	}

	if pkg.BuildDir == "" {
		pkg.BuildDir = pkgGetBuildDirFromViper(pkg.Name)
	}

	if pkg.InstallDir == "" {
		pkg.InstallDir = pkgGetInstallDirFromViper(pkg.Name)
	}

	if pkg.SourceDir == "" {
		pkg.SourceDir = pkgGetSourceDirFromViper(pkg.Name)
	}

	if pkg.TargetName == "" {
		pkg.TargetName = pkg.Name
	}

	if pkg.Url == "" {
		pkg.Url = pkg.Git
	}

	if pkg.Url == "" {
		pkg.Url, err = pkgGetUrlFromViper(pkg.Name)
	}

	if pkg.BuildSystem == "" {
		pkg.BuildSystem = pkgGetBuildSystemFromViper(pkg.Name)
	}

	if len(pkg.EnvVars) == 0 {
		pkg.EnvVars, err = pkgGetEnvVarsFromViper(pkg.Name)
		if err != nil {
			return err
		}
	}

	if len(pkg.LdLibraryPath) == 0 {
		pkg.LdLibraryPath = append(pkg.LdLibraryPath,
			pkgGetLdLibraryPathFromViper(pkg.Name)...)
	}

	if pkg.BuildType == "" {
		pkg.BuildType = easifem_build_type
	}

	if pkg.BuildSharedLibs && pkg.BuildStaticLibs {
		pkg.BuildSharedLibs = true
	}

	return err
}

//----------------------------------------------------------------------------
//                                                          PkgMakeFromViper
//----------------------------------------------------------------------------

// Make pkg from Viper config
func PkgMakeFromViper(name string) (*Pkg, error) {
	p := &Pkg{IsActive: true, Name: name, IsExtPkg: false}
	err := PkgCheckAndFix(p)
	return p, err
}

//----------------------------------------------------------------------------
//                                                            PkgDownloadPkg
//----------------------------------------------------------------------------

// Download the package
func pkgDownloadPkg(url, source_dir, pwd string) error {
	return pkgDownloadPkgWithOutput(url, source_dir, pwd, nil)
}

func pkgDownloadPkgWithOutput(url, source_dir, pwd string, operation *libraryOutput) (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var display *stageOutput
	if operation != nil {
		display = operation.begin("Download")
	}
	shared := display != nil
	if !shared {
		display = newStageOutput(os.Stderr, "Download", animatedOutput(), quiet)
	}
	defer func() {
		if shared {
			operation.complete("Download", err)
		} else {
			display.finish(err)
		}
	}()
	req := &getter.Request{
		Src: url, Dst: source_dir, Pwd: pwd,
		GetMode: getter.ModeAny, Copy: true, DisableSymlinks: true,
		ProgressListener: &ProgressBar{stage: display},
	}
	client := *getter.DefaultClient
	client.DisableSymlinks = true
	if _, err = client.Get(ctx, req); err != nil {
		return fmt.Errorf("download %s: %w", source_dir, err)
	}
	return nil
}

//----------------------------------------------------------------------------
//                                                                 pkgMakeDir
//----------------------------------------------------------------------------

// make directory for installing and building pkg
func pkgMakeDir(dir string) {
	if err := makePackageDirs(dir); err != nil {
		log.Fatal(err)
	}
}

// Library operations return setup failures so the shared dashboard can restore
// the cursor through its deferred cleanup.
func makePackageDirs(dir string) error {
	for _, opt := range []string{"lib", "include", "bin", "share"} {
		astr := path.Join(dir, opt)
		if err := os.MkdirAll(astr, 0777); err != nil {
			return fmt.Errorf("create directory %s: %w", astr, err)
		}
	}
	return nil
}

//----------------------------------------------------------------------------
//                                                            pkgGetEnvVars
//----------------------------------------------------------------------------

// Get LD_LIBRARY_PATH from viper
func pkgGetLdLibraryPathFromViper(pkg string) []string {
	const KEY = ".ldLibraryPath"
	var ans []string

	if key := pkg + KEY; viper.IsSet(key) {
		ans = viper.GetStringSlice(key)
	} else if key := easifem_current_env_name + KEY; viper.IsSet(key) {
		ans = viper.GetStringSlice(key)
	}

	return ans
}

//----------------------------------------------------------------------------
//                                                            pkgGetEnvVars
//----------------------------------------------------------------------------

func pkgGetEnvVarsFromViper(pkg string) (map[string]string, error) {
	ans := make(map[string]string)

	strs := []string{}

	if key := pkg + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	} else if key := easifem_current_env_name + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	}

	for _, s := range strs {
		key, value, ok := strings.Cut(s, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("package %s: invalid environment entry %q; expected KEY=value", pkg, s)
		}
		ans[key] = value
	}

	return ans, nil
}

//----------------------------------------------------------------------------
//                                                pkgGetBuildSystemFromViper
//----------------------------------------------------------------------------

// This function reads buildSystem from [pkg]
// If failed to read it tries to read from env.buildSystem
// If failed it returns "cmake"
func pkgGetBuildSystemFromViper(name string) string {
	if key := name + ".buildSystem"; viper.IsSet(key) {
		return viper.GetString(key)
	}

	if key := easifem_current_env_name + ".buildSystem"; viper.IsSet(key) {
		return viper.GetString(key)
	}

	return easifem_build_system
}

//----------------------------------------------------------------------------
//                                                    pkgGetSourceDirFromViper
//----------------------------------------------------------------------------

// this function returns the source directory name from viper config
// First get from pkg.sourceDir
// get sourceDir from the viper
// it uses the default value sourceDir/easifem/pkg
func pkgGetSourceDirFromViper(pkg string) string {
	if key := pkg + ".sourceDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := easifem_current_env_name + ".sourceDir"; viper.IsSet(key) {
		return path.Join(viper.GetString(key), pkg)
	}
	return path.Join(sourceDir, "easifem", pkg)
}

//----------------------------------------------------------------------------
//                                                    pkgGetBuildDirFromViper
//----------------------------------------------------------------------------

// this function returns the build directory name from viper config
// First get pkg.buildDir
// If failed to get it tries to get env.buildDir
// If failed use the default value buildDir/easifem/pkg
func pkgGetBuildDirFromViper(pkg string) string {
	if key := pkg + ".buildDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := easifem_current_env_name + ".buildDir"; viper.IsSet(key) {
		return path.Join(viper.GetString(key), pkg)
	}
	return path.Join(buildDir, "easifem", pkg)
}

//----------------------------------------------------------------------------
//                                                  pkgGetInstallDirFromViper
//----------------------------------------------------------------------------

// this function returns the install directory name from viper config
// First get pkg.installDir
// If failed to get it tries to get env.installDir
// If failed use the default value installDir/easifem/pkg
func pkgGetInstallDirFromViper(pkg string) string {
	if key := pkg + ".installDir"; viper.IsSet(key) {
		return viper.GetString(key)
	}
	if key := easifem_current_env_name + ".installDir"; viper.IsSet(key) {
		return path.Join(viper.GetString(key), pkg)
	}
	return path.Join(installDir, "easifem", pkg)
}

//----------------------------------------------------------------------------
//                                                        pkgGetUrlFromViper
//----------------------------------------------------------------------------

// this function returns the url from config file
func pkgGetUrlFromViper(pkg string) (string, error) {
	opts := []string{
		"git",
		"url",
	}

	for _, opt := range opts {
		if key := pkg + "." + opt; viper.IsSet(key) {
			return viper.GetString(key), nil
		}
	}
	return "", errors.New("no url related tag found")
}

//----------------------------------------------------------------------------
//                                                           printLogFromToml
//----------------------------------------------------------------------------

// print the log from toml
// func printLogFromToml(p Pkg, meta toml.MetaData) {
// 	log.Printf("[log] :: pkg.go | printLogFromToml() pkg.Name %s, Decoded", p.Name)
//
// 	indent := "[log] :: pkg.go | printLogFromToml() pkg." + p.Name + "."
//
// 	typ, val := reflect.TypeOf(p), reflect.ValueOf(p)
// 	for i := 0; i < typ.NumField(); i++ {
// 		log.Printf("%s%-11s → %v\n", indent, typ.Field(i).Name, val.Field(i).Interface())
// 	}
//
// 	log.Printf("%s Keys", indent)
// 	keys := meta.Keys()
// 	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
// 	for _, k := range keys {
// 		log.Printf("%s%-10s %s\n", indent, meta.Type(k...), k)
// 	}
//
// 	log.Printf("%s Undecoded", indent)
// 	keys = meta.Undecoded()
// 	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
// 	for _, k := range keys {
// 		log.Printf("%s%-10s %s\n", indent, meta.Type(k...), k)
// 	}
// }

//----------------------------------------------------------------------------
//                                                                cmakeOnOff
//----------------------------------------------------------------------------

func cmakeOnOff(a bool) string {
	if a {
		return "ON"
	} else {
		return "OFF"
	}
}

//----------------------------------------------------------------------------
//                                                                pkgRunCmd
//----------------------------------------------------------------------------

func pkgRunCmd(cargs []string, pkg, step string) error {
	return runBuildStage(cargs, pkg, strings.Trim(step, "[]"), "")
}

//----------------------------------------------------------------------------
//                                                             pkgGetExtPkgs
//----------------------------------------------------------------------------

// This function reads env.extpkgs from the viper if it exists
// otherwise it reads all the pkgs in the cofig directory's extpkgs
// If config directory is not available, then it returns the default external packages
func pkgGetExtPkgs() []string {
	if key := easifem_current_env_name + ".extpkgs"; viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}

	var names []string
	for name, pkg := range easifem_pkgs {
		if pkg.IsActive && pkg.IsExtPkg {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

//----------------------------------------------------------------------------
//                                                             pkgAllPkgNames
//----------------------------------------------------------------------------

// This function reads env.extpkgs from the viper if it exists
// otherwise it reads all the pkgs in the cofig directory's extpkgs
// If config directory is not available, then it returns the default external packages
func pkgGetAllNames() []string {
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		mypath := path.Join(configPath, easifem_pkg_config_dir)
		entries, err2 := os.ReadDir(mypath)
		if err2 != nil {
			log.Fatalln("[err] :: pkg.go | pkgGetExtPkgs() ➡️ ", err2)
		}

		ans := make([]string, len(entries))
		const suffix = ".toml"

		for i, e := range entries {
			ans[i] = strings.TrimSuffix(e.Name(), suffix)
		}

		return ans

	}

	return make([]string, 0)
}

//----------------------------------------------------------------------------
//                                                      makeAllPkgsFromToml
//----------------------------------------------------------------------------

// This function is called in the root command

func makeAllPkgsFromToml() error {
	// read all packages from toml file
	pkgs := pkgGetAllNames()

	var err error
	var p *Pkg

	for _, pkg := range pkgs {
		p, err = PkgMakeFromToml(pkg)
		if err != nil {
			err = fmt.Errorf("makeAllPkgsFromToml() pkg=%s, err=%w", pkg, err)
			break
		}

		easifem_pkgs[p.Name] = p

	}

	return err
}

// Read the cache
func readCache() error {
	var err error
	cacheFile := path.Join(configPath, easifem_cache_file)

	if verbose && !quiet {
		log.Printf("[log] :: pkg.go | readCache() | cacheFile=%s➡️ ", cacheFile)
	}

	_, err = toml.DecodeFile(cacheFile, &easifem_cache)
	if err != nil {
		return err
	}

	return err
}

// Build in memory cache
func buildCache() {
	for _, pkg := range easifem_pkgs {
		if pkg.IsActive {
			easifem_cache.INSTALL_DIRS = append(easifem_cache.INSTALL_DIRS, pkg.InstallDir)

			for _, prefixpath := range pkg.CmakePrefixPath {
				easifem_cache.INSTALL_DIRS = append(easifem_cache.INSTALL_DIRS,
					os.ExpandEnv(prefixpath))
			}

			// easifem_cache.INSTALL_DIRS = append(easifem_cache.INSTALL_DIRS, pkg.CmakePrefixPath...)
			easifem_cache.LD_LIBRARY_PATH = append(easifem_cache.LD_LIBRARY_PATH, path.Join(pkg.InstallDir, "lib"))
			easifem_cache.LD_LIBRARY_PATH = append(easifem_cache.LD_LIBRARY_PATH, pkg.LdLibraryPath...)
		}
	}
}

// Read the cache
func writeCache() error {
	cacheFile := path.Join(configPath, easifem_cache_file)

	f, err := os.Create(cacheFile)
	if err != nil {
		return fmt.Errorf("writeCache() | error = %w", err)
	}

	if verbose && !quiet {
		log.Printf("[log] :: pkg.go | writeCache() | cacheFile=%s➡️ ", cacheFile)
	}

	err = toml.NewEncoder(f).Encode(easifem_cache)
	if err != nil {
		return fmt.Errorf("writeCache() | error = %w", err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("writeCache() | error = %w", err)
	}

	return err
}

//----------------------------------------------------------------------------
//                                                          writeShellVarFle
//----------------------------------------------------------------------------

// Write easifemvar.sh and easifemvar.fish

func writeShellVarFle() {
	// open output file
	fo, err := os.Create(path.Join(configPath, easifem_shell_var_file+".fish"))
	if err != nil {
		log.Fatalln("[err] :: pkg.go | writeShellVarFle() | err while opening file => ", err)
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			log.Fatalln("[err] :: pkg.go | writeShellVarFle() | err while closing file => ", err)
		}
	}()

	// make a buffer to keep chunks that are read
	for _, val := range easifem_cache.LD_LIBRARY_PATH {
		// write a chunk
		aline := []byte("set -gx LD_LIBRARY_PATH " + val + " $LD_LIBRARY_PATH\n")
		if _, err := fo.Write(aline); err != nil {
			log.Fatalln("[err] :: pkg.go | writeShellVarFle() | err while writing to file => ", err)
		}
	}
}
