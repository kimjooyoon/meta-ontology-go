package bodyexecution

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGoModuleCacheUsesOneConfiguredOrDefaultLocation(t *testing.T) {
	root := t.TempDir()
	configured, goPath := filepath.Join(root, "modules"), filepath.Join(root, "workspace")
	for _, tc := range []struct{ configured, goPath, home, want string }{
		{configured, goPath, root, configured},
		{"", goPath, root, filepath.Join(goPath, "pkg", "mod")},
		{"", goPath + string(os.PathListSeparator) + filepath.Join(root, "second"), root, filepath.Join(goPath, "pkg", "mod")},
		{"", "", root, filepath.Join(root, "go", "pkg", "mod")},
		{"relative-cache", goPath, root, ""},
		{"", "relative-workspace", root, ""},
		{"", "", "", ""},
	} {
		if got := goModuleCache(tc.configured, tc.goPath, tc.home); got != tc.want {
			t.Fatalf("cache(%+v) = %q", tc, got)
		}
	}
}

func TestTrimmedCompilerResolvesOnlyTheExactCachedTool(t *testing.T) {
	cache := t.TempDir()
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	cached := filepath.Join(cache, "golang.org", "toolchain@v0.0.1-go1.27.1."+runtime.GOOS+"-"+runtime.GOARCH, "bin", name)
	for _, pathAvailable := range []bool{false, true} {
		for _, cacheSupported := range []bool{false, true} {
			pathTool := filepath.Join(cache, "old-path-go")
			var inspected []string
			lookup := func(input string) (string, error) {
				if input == "go" && pathAvailable {
					return pathTool, nil
				}
				if input == cached {
					return cached, nil
				}
				return "", os.ErrNotExist
			}
			check := func(path string) bool {
				inspected = append(inspected, path)
				return path == cached && cacheSupported
			}
			got, origin, err := resolveGoTool("", "", cache, lookup, check)
			switch {
			case cacheSupported:
				if err != nil || got != cached || origin != "toolchain_cache" {
					t.Fatal("trimmed build lost matching cache", got, origin, err)
				}
			case pathAvailable:
				if err != nil || got != pathTool || origin != "path_fallback" {
					t.Fatal("unsupported cache replaced PATH observation", got, origin, err)
				}
			default:
				if err == nil || got != "" || origin != "" {
					t.Fatal("missing local tool accepted", got, origin, err)
				}
			}
			if len(inspected) > 2 || inspected[len(inspected)-1] != cached {
				t.Fatal("cache search exceeded the exact version/platform location", inspected)
			}
		}
	}
}
