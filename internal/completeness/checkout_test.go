package completeness

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitCheckoutPreservesDeclaredReceiptAndBinaryBytes(t *testing.T) {
	// CI materializes its workspace through Git environment overrides.
	outerGit := filepath.Join(t.TempDir(), "outer.git")
	outerTree := t.TempDir()
	outerIndex := filepath.Join(t.TempDir(), "outer.index")
	t.Setenv("GIT_DIR", outerGit)
	t.Setenv("GIT_WORK_TREE", outerTree)
	t.Setenv("GIT_INDEX_FILE", outerIndex)
	attributes, err := os.ReadFile("../../.gitattributes")
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"internal/completeness/receipt.gooo": DeclarationSource(),
		"examples/body.gooo.fixture":         []byte("package sample\nnamespace sample\n"),
		"models/weights.bin":                 {0, 1, '\r', '\n', 2, '\n', 0xff},
	}
	for _, name := range []string{"receipt.generated.go", "receipt.schema.json"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files["internal/completeness/"+name] = data
	}
	for _, tc := range []struct {
		name       string
		attributes []byte
		stable     bool
	}{
		{"repository_attributes", attributes, true},
		{"git_default_conversion", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			repository, checkout := filepath.Join(root, "repository"), filepath.Join(root, "checkout")
			if err := os.MkdirAll(repository, 0o755); err != nil {
				t.Fatal(err)
			}
			for name, data := range files {
				path := filepath.Join(repository, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(repository, ".gitattributes"), tc.attributes, 0o644); err != nil {
				t.Fatal(err)
			}
			git := func(args ...string) {
				t.Helper()
				args = append([]string{"-c", "core.autocrlf=true", "-c", "core.safecrlf=false"}, args...)
				command := exec.Command("git", args...)
				command.Dir = repository
				command.Env = checkoutGitEnvironment()
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v: %s", args, err, output)
				}
			}
			git("init", "--quiet")
			git("add", "--", ".")
			git("checkout-index", "--all", "--prefix="+filepath.ToSlash(checkout)+"/")
			for name, want := range files {
				got, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
				if err != nil {
					t.Fatal(err)
				}
				stable := tc.stable || name == "models/weights.bin"
				if bytes.Equal(got, want) != stable {
					t.Fatalf("checkout byte preservation for %s: want stable=%v", name, stable)
				}
			}
		})
	}
	for _, path := range []string{outerGit, outerIndex} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("fixture wrote to inherited Git path %s: %v", path, err)
		}
	}
	entries, err := os.ReadDir(outerTree)
	if err != nil || len(entries) != 0 {
		t.Fatalf("fixture changed inherited work tree: entries=%d error=%v", len(entries), err)
	}
}

func checkoutGitEnvironment() []string {
	environment := make([]string, 0, len(os.Environ())+2)
	for _, variable := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(variable), "GIT_") {
			environment = append(environment, variable)
		}
	}
	return append(environment, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
}
