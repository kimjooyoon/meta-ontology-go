package completeness

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitCheckoutPreservesDeclaredReceiptAndBinaryBytes(t *testing.T) {
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
				command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
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
}
