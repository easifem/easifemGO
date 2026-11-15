// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// writePlugin creates plugins/<name>.toml under a temporary configPath.
func writePlugin(t *testing.T, name, body string) {
	t.Helper()
	oldPath, oldEnv := configPath, easifem_current_env_name
	t.Cleanup(func() { configPath, easifem_current_env_name = oldPath, oldEnv; viper.Reset() })
	viper.Reset()
	configPath = t.TempDir()
	easifem_current_env_name = "dev"
	dir := filepath.Join(configPath, easifem_pkg_config_dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildSettingsFallBackToEnvironment(t *testing.T) {
	writePlugin(t, "base", "name = \"base\"\n")
	viper.Set("dev.buildType", "Debug")
	viper.Set("dev.buildSharedLibs", true)
	viper.Set("dev.buildStaticLibs", true)
	viper.Set("release.buildType", "Release")

	p, err := PkgMakeFromToml("base")
	if err != nil {
		t.Fatal(err)
	}
	if p.BuildType != "Debug" || !p.BuildSharedLibs || !p.BuildStaticLibs {
		t.Fatalf("got buildType=%q shared=%v static=%v, want the [dev] values", p.BuildType, p.BuildSharedLibs, p.BuildStaticLibs)
	}
}

func TestPluginBuildSettingsOverrideEnvironment(t *testing.T) {
	// Explicit false in the plugin must win over true in the environment.
	writePlugin(t, "base", "name = \"base\"\nbuildType = \"Release\"\nBuildSharedLibs = false\nbuildStaticLibs = false\n")
	viper.Set("dev.buildType", "Debug")
	viper.Set("dev.buildSharedLibs", true)
	viper.Set("dev.buildStaticLibs", true)

	p, err := PkgMakeFromToml("base")
	if err != nil {
		t.Fatal(err)
	}
	if p.BuildType != "Release" || p.BuildSharedLibs || p.BuildStaticLibs {
		t.Fatalf("got buildType=%q shared=%v static=%v, want the plugin values", p.BuildType, p.BuildSharedLibs, p.BuildStaticLibs)
	}
}

func TestPackageTableBeatsEnvironmentAndDefault(t *testing.T) {
	writePlugin(t, "base", "name = \"base\"\n")
	viper.Set("base.buildType", "RelWithDebInfo")
	viper.Set("dev.buildType", "Debug")

	p, err := PkgMakeFromToml("base")
	if err != nil {
		t.Fatal(err)
	}
	if p.BuildType != "RelWithDebInfo" {
		t.Fatalf("buildType = %q, want the [base] value", p.BuildType)
	}

	viper.Reset()
	p, err = PkgMakeFromToml("base")
	if err != nil {
		t.Fatal(err)
	}
	if p.BuildType != easifem_build_type || p.BuildSharedLibs || p.BuildStaticLibs {
		t.Fatalf("got buildType=%q shared=%v static=%v, want the defaults", p.BuildType, p.BuildSharedLibs, p.BuildStaticLibs)
	}
}
