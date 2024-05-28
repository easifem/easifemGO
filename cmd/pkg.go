package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/briandowns/spinner"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/spf13/viper"
)

//----------------------------------------------------------------------------
//                                                                    PkgMake
//----------------------------------------------------------------------------

func PkgMake(pkg *Pkg, pwd string) {
	change_dir(pkg.BuildDir)

	cargs := []string{
		path.Join(pkg.SourceDir, "configure"),
		"--prefix=" + pkg.InstallDir,
	}

	var err error

	cargs = append(cargs, pkg.BuildOptions...)
	err = pkgRunCmd(cargs, pkg.Name, "[config]")
	if err != nil {
		log.Fatalf("[err] :: PkgMake() | config step failed %v", err)
	}

	cargs = []string{"make"}
	err = pkgRunCmd(cargs, pkg.Name, "[build]")
	if err != nil {
		log.Fatalf("[err] :: PkgMake() | build step failed %v", err)
	}

	cargs = []string{"make", "install"}
	err = pkgRunCmd(cargs, pkg.Name, "[install]")
	if err != nil {
		log.Fatalf("[err] :: PkgMake() | install step failed %v", err)
	}

	change_dir(pwd)
}

//----------------------------------------------------------------------------
//                                                              PkgCMake
//----------------------------------------------------------------------------

func PkgCmake(pkg *Pkg, pwd string) {
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
	err = pkgRunCmd(cargs, pkg.Name, "[config]")
	if err != nil {
		log.Fatalf("[err] :: pkg.go | PkgCmake() | config step failed %v", err)
	}

	cargs = []string{"cmake", "--build", pkg.BuildDir}
	err = pkgRunCmd(cargs, pkg.Name, "[build]")
	if err != nil {
		log.Fatalf("[err] :: pkg.go | PkgCmake() | build step failed %v", err)
	}

	cargs = []string{"cmake", "--install", pkg.BuildDir}
	err = pkgRunCmd(cargs, pkg.Name, "[install]")
	if err != nil {
		log.Fatalf("[err] :: pkg.go | PkgCmake() | install step failed %v", err)
	}
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

func PkgInstall(pkg *Pkg, pwd string) error {
	if !pkg.IsActive {
		return nil
	}

	if !noDownload {
		pkgDownloadPkg(pkg.Url, pkg.SourceDir, pwd)
	}

	pkgMakeDir(pkg.InstallDir)

	for k, v := range pkg.EnvVars {
		if !quiet {
			log.Printf("[log] :: install.go | setting env %s=%s", k, v)
		}
		os.Setenv(k, v)
	}

	switch pkg.BuildSystem {
	case "make":
		PkgMake(pkg, pwd)
	case "cmake":
		PkgCmake(pkg, pwd)
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

	if !quiet {
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
		pkg.EnvVars = pkgGetEnvVarsFromViper(pkg.Name)
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
func pkgDownloadPkg(url, source_dir, pwd string) {
	ctx, cancel := context.WithCancel(context.Background())
	// Build the client
	req := &getter.Request{
		Src:             url,
		Dst:             source_dir,
		Pwd:             pwd,
		GetMode:         getter.ModeAny,
		Copy:            true,
		DisableSymlinks: true,
	}
	req.ProgressListener = defaultProgressBar
	wg := sync.WaitGroup{}
	wg.Add(1)

	client := getter.DefaultClient
	client.DisableSymlinks = true

	getters := getter.Getters
	client.Getters = getters

	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		res, err := client.Get(ctx, req)
		if err != nil {
			errChan <- err
			log.Fatalf("[err] :: pkgDownloadPkg.go: %s", res.Dst)
			return
		}
		if !quiet {
			log.Printf("[log] :: pkgDownloadPkg.go | source:  %s", req.Src)
			log.Printf("[log] :: pkgDownloadPkg.go | request destination:  %s", req.Dst)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		if !quiet {
			log.Printf("[log] :: pkgDownloadPkg.go | signal: %v", sig)
		}

	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		log.Fatalf("[err] :: pkgDownloadPkg.go | Error downloading: %s", err)
	}
}

//----------------------------------------------------------------------------
//                                                                 pkgMakeDir
//----------------------------------------------------------------------------

// make directory for installing and building pkg
func pkgMakeDir(dir string) {
	for _, opt := range []string{"lib", "include", "bin", "share"} {
		astr := path.Join(dir, opt)
		if err := os.MkdirAll(astr, 0777); err != nil {
			log.Fatalln("[err] :: make_dir.go | cannot create dir ➡️  ", astr,
				" with permission 0777")
		}
	}
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

func pkgGetEnvVarsFromViper(pkg string) map[string]string {
	ans := make(map[string]string)

	strs := []string{}

	if key := pkg + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	} else if key := easifem_current_env_name + ".env"; viper.IsSet(key) {
		strs = viper.GetStringSlice(key)
	}

	for _, s := range strs {
		kv := strings.Split(s, "=")
		ans[kv[0]] = kv[1]
	}

	return ans
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
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " installing " + pkg + " cmd: " + cargs[0] + " " + step
	_ = s.Color("red")
	if quiet {
		s.Start()
		defer s.Stop()
	}

	if !quiet {
		log.Println("[log] :: pkgRunCmd.go | cmd name ➡️ ", cargs)
	}
	cmd := exec.Command(cargs[0], cargs[1:]...)

	output_stdout := make(chan []byte)

	var err error
	go pkgExecuteCmd(cmd, output_stdout, err)

	if !quiet {
		for data := range output_stdout {
			log.Println(string(data))
		}
	} else {
		for range output_stdout {
		}
	}

	return err
}

//----------------------------------------------------------------------------
//                                                              pkgExecuteCmd
//----------------------------------------------------------------------------

func pkgExecuteCmd(cmd *exec.Cmd, output_stdout chan []byte, err error) {
	defer close(output_stdout)
	var stdout io.ReadCloser

	if err != nil {
		output_stdout <- []byte(fmt.Sprintf("Error before executing devExecuteCmd(): %v", err))
		return
	}

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		output_stdout <- []byte(fmt.Sprintf("Error getting stdout pipe: %v", err))
		return
	}

	cmd.Stderr = cmd.Stdout
	stdout_scanner := bufio.NewScanner(stdout)

	done := make(chan struct{})

	err = cmd.Start()
	if err != nil {
		output_stdout <- []byte(fmt.Sprintf("Error executing: %v", err))
		return
	}

	go func() {
		for stdout_scanner.Scan() {
			output_stdout <- stdout_scanner.Bytes()
		}
		done <- struct{}{}
	}()

	<-done

	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
		output_stdout <- []byte(fmt.Sprintf("Error waiting for the script to complete: %v", err))
	}
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

	return []string{"sparsekit", "lapack95", "fftw", "superlu", "arpack", "tomlf", "lis"}
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

	if !quiet {
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

	if !quiet {
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
