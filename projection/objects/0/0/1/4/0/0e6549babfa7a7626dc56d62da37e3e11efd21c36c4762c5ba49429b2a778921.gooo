package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckSourcePolicyRejectsMixedDirectoryEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mixed", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mixed", "file.go"), []byte("package mixed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mixed", "nested", "a.go"), []byte("package mixed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	policy := DefaultLinePolicy()
	if err := CheckSourcePolicy(root, nil, policy); err == nil {
		t.Fatal("mixed directory entries were accepted")
	}
}
func TestCheckSourcePolicyRejectsTooManyDirectEntries(t *testing.T) {
	root := t.TempDir()
	for i := 1; i <= 11; i++ {
		path := filepath.Join(root, "many", "f"+fmt.Sprint(i)+".go")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package many\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	policy := DefaultLinePolicy()
	if err := CheckSourcePolicy(root, nil, policy); err == nil {
		t.Fatal("too many direct entries were accepted")
	}
}
