package bodyexecution

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func selectGoTool(requested string) (string, string, error) {
	return resolveGoTool(requested, runtime.GOROOT(), localGoModuleCache(), exec.LookPath, nativeGoBuildInfo)
}

func localGoModuleCache() string {
	home, _ := os.UserHomeDir()
	return goModuleCache(os.Getenv("GOMODCACHE"), os.Getenv("GOPATH"), home)
}

func goModuleCache(configured, goPath, home string) string {
	if configured != "" {
		if filepath.IsAbs(configured) {
			return configured
		}
		return ""
	}
	if goPath == "" && home != "" {
		goPath = filepath.Join(home, "go")
	}
	paths := filepath.SplitList(goPath)
	if len(paths) == 0 || !filepath.IsAbs(paths[0]) {
		return ""
	}
	return filepath.Join(paths[0], "pkg", "mod")
}

// Defaults inspect three local locations in a fixed order. Selection starts no
// child and downloads nothing; the existing executor still binds current bytes
// and observes the selected tool's actual version before building a body.
func resolveGoTool(requested, goRoot, moduleCache string, lookup func(string) (string, error),
	supported func(string) bool) (string, string, error) {
	absolute := func(path, origin string, err error) (string, string, error) {
		if err != nil {
			return "", "", fmt.Errorf("Go tool is unavailable: %w; select an installed Go 1.27.1 executable with --go-bin", err)
		}
		path, err = filepath.Abs(path)
		return path, origin, err
	}
	if requested != "" {
		path, err := lookup(requested)
		return absolute(path, "explicit", err)
	}
	path, pathErr := lookup("go")
	if pathErr == nil && supported(path) {
		return absolute(path, "path", nil)
	}
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if goRoot != "" {
		rootTool, err := lookup(filepath.Join(goRoot, "bin", name))
		if err == nil && rootTool != path && supported(rootTool) {
			return absolute(rootTool, "compiler_goroot", nil)
		}
	}
	if moduleCache != "" {
		// Go's toolchain module has one version/platform-specific location.
		// This remains usable when -trimpath omitted the compiler's GOROOT.
		version := "golang.org/toolchain@v0.0.1-go1.27.1." + runtime.GOOS + "-" + runtime.GOARCH
		cachedTool, err := lookup(filepath.Join(moduleCache, version, "bin", name))
		if err == nil && cachedTool != path && supported(cachedTool) {
			return absolute(cachedTool, "toolchain_cache", nil)
		}
	}
	// Preserve a PATH wrapper or wrong version as an actual observed failure
	// when the compiler's local native tool cannot be found.
	return absolute(path, "path_fallback", pathErr)
}
