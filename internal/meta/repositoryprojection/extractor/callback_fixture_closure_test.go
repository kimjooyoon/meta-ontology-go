package extractor

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCallbackFixtureSelectsRootPackagesAndCandidateEmbeds(t *testing.T) {
	root := t.TempDir()
	const logical = "cmd/fixture/fixture_test.go"
	const source = `package fixture
import (
    "testing"
    "example.test/callbackfixture/origin"
)
func TestValue(t *testing.T) {
    if origin.Value != "ready" { t.Fatal(origin.Value) }
}
`
	for name, raw := range map[string]string{
		"go.mod":                "module example.test/callbackfixture\n\ngo 1.27.3\n",
		logical:                 source,
		"origin/value.go":       "package origin\nimport _ \"embed\"\n//go:embed value.txt\nvar Value string\n",
		"origin/value.txt":      "ready",
		"replacement/value.go":  "package replacement\nimport _ \"embed\"\n//go:embed value.txt\nvar Value string\n",
		"replacement/value.txt": "ready",
		"unused/broken.go":      "package unused\nfunc Broken() { missing() }\n",
	} {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(raw), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct{ name, candidate, dependency string }{
		{"original", "", "origin"},
		{"candidate", strings.ReplaceAll(source, "origin", "replacement"), "replacement"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			output := t.TempDir()
			if err := copyCallbackPreviewModuleInputs(ctx, root, output, logical, test.candidate); err != nil {
				t.Fatal(err)
			}
			assertCallbackFixtureInputs(t, output, logical, test.dependency)
			command := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "./cmd/fixture", "-run", "^TestValue$")
			command.Dir, command.Env = output, callbackObservationEnvironment()
			if raw, err := command.CombinedOutput(); err != nil {
				t.Fatalf("selected fixture cannot execute: %v\n%s", err, raw)
			}
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
			if err != nil || !bytes.Equal(raw, []byte(source)) {
				t.Fatalf("candidate overlay changed authored source: %q, %v", raw, err)
			}
		})
	}
}

func assertCallbackFixtureInputs(t *testing.T, root, logical, dependency string) {
	t.Helper()
	unselected := "origin"
	if dependency == "origin" {
		unselected = "replacement"
	}
	for _, name := range []string{"go.mod", logical, dependency + "/value.go", dependency + "/value.txt"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			t.Fatalf("selected source or embed missing: %s: %v", name, err)
		}
	}
	for _, name := range []string{"go.sum", "unused/broken.go", unselected + "/value.go"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); !os.IsNotExist(err) {
			t.Fatalf("unselected input was materialized: %s: %v", name, err)
		}
	}
}
