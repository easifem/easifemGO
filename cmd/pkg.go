package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/briandowns/spinner"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/spf13/viper"
)

const pkgConfigDir = "plugins"

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
	License         string
	Name            string
	RunTest         bool
	SourceDir       string
	TargetName      string
	Url             string
}

//----------------------------------------------------------------------------
//                                                                   PkgMake
//----------------------------------------------------------------------------

func PkgMake(pkg *Pkg, pwd string) {
	change_dir(pkg.BuildDir)

	cargs := []string{
		path.Join(pkg.SourceDir, "configure"),
		"--prefix=" + pkg.InstallDir,
	}
	cargs = append(cargs, pkg.BuildOptions...)
	pkgRunCmd(cargs, pkg.Name, "[config]")

	cargs = []string{"make"}
	pkgRunCmd(cargs, pkg.Name, "[build]")

	cargs = []string{"make", "install"}
	pkgRunCmd(cargs, pkg.Name, "[install]")

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

	if len(pkg.CmakePrefixPath) != 0 {
		cargs = append(cargs, "-D CMAKE_PREFIX_PATH:PATH="+strings.Join(pkg.CmakePrefixPath, ";"))
	}

	cargs = append(cargs, pkg.BuildOptions...)
	pkgRunCmd(cargs, pkg.Name, "[config]")

	cargs = []string{"cmake", "--build", pkg.BuildDir}
	pkgRunCmd(cargs, pkg.Name, "[build]")

	cargs = []string{"cmake", "--install", pkg.BuildDir}
	pkgRunCmd(cargs, pkg.Name, "[install]")
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

	pkgDownloadPkg(pkg.Url, pkg.SourceDir, pwd)
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
func PkgMakeFromToml(tomlFile string) (*Pkg, error) {
	var p *Pkg

	if !quiet {
		log.Println("[log] :: pkg.go | MakePkgFromToml() ➡️ ", tomlFile)
	}

	p = &Pkg{IsActive: true, IsExtPkg: false}
	// meta, err := toml.DecodeFile(tomlFile, &p)
	_, err := toml.DecodeFile(tomlFile, p)
	if err != nil {
		return p, err
	}

	if err := PkgCheckAndFix(p); err != nil {
		return p, err
	}

	return p, nil
}

//----------------------------------------------------------------------------
//                                                                PkgLogPrint
//----------------------------------------------------------------------------

// This function prints the pkg by using log.Println
func PkgLogPrint(p *Pkg) {
	log.Println("[log] :: pkg.go | PkgLogPrint() ➡️ ")
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

	if pkg.BuildType == "" {
		pkg.BuildType = "Release"
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

	return "cmake"
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
func printLogFromToml(p Pkg, meta toml.MetaData) {
	log.Printf("[log] :: pkg.go | printLogFromToml() pkg.Name %s, Decoded", p.Name)

	indent := "[log] :: pkg.go | printLogFromToml() pkg." + p.Name + "."

	typ, val := reflect.TypeOf(p), reflect.ValueOf(p)
	for i := 0; i < typ.NumField(); i++ {
		log.Printf("%s%-11s → %v\n", indent, typ.Field(i).Name, val.Field(i).Interface())
	}

	log.Printf("%s Keys", indent)
	keys := meta.Keys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, k := range keys {
		log.Printf("%s%-10s %s\n", indent, meta.Type(k...), k)
	}

	log.Printf("%s Undecoded", indent)
	keys = meta.Undecoded()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, k := range keys {
		log.Printf("%s%-10s %s\n", indent, meta.Type(k...), k)
	}
}

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

func pkgRunCmd(cargs []string, pkg, step string) {
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

	go pkgExecuteCmd(cmd, output_stdout)

	if !quiet {
		for data := range output_stdout {
			log.Println(string(data))
		}
	} else {
		for range output_stdout {
		}
	}
}

//----------------------------------------------------------------------------
//                                                              pkgExecuteCmd
//----------------------------------------------------------------------------

func pkgExecuteCmd(cmd *exec.Cmd, output_stdout chan []byte) {
	defer close(output_stdout)
	stdout, err := cmd.StdoutPipe()
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
		mypath := path.Join(configPath, pkgConfigDir)
		entries, err2 := os.ReadDir(mypath)
		if err2 != nil {
			log.Fatalln("[err] :: get_ext_pkgs.go | get_ext_pkgs() ➡️ ", err2)
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
