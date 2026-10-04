package bodyexecution

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func selectGoTool(requested string) (string, string, error) {
	return resolveGoTool(requested, runtime.GOROOT(), exec.LookPath, nativeGoBuildInfo)
}

// Defaults inspect two local locations in a fixed order. Selection starts no
// child and downloads nothing; the existing executor still binds current bytes
// and observes the selected tool's actual version before building a body.
func resolveGoTool(requested, goRoot string, lookup func(string) (string, error),
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
	// Preserve a PATH wrapper or wrong version as an actual observed failure
	// when the compiler's local native tool cannot be found.
	return absolute(path, "path_fallback", pathErr)
}
