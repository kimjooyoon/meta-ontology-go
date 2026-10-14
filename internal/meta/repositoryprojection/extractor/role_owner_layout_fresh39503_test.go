package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRoleLayout39503(t *testing.T, split bool) (string, map[string]bool) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "internal/meta/syntaxregistration")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	required := map[string]bool{}
	whole := "package syntaxregistration\n"
	for i, name := range []string{"RoleAlpha39503", "RoleBeta39503", "RoleGamma39503", "RoleDelta39503"} {
		required[name] = true
		body := fmt.Sprintf("func %s() int { return %d }\n", name, 39503+i)
		if split {
			writeRoleFile39503(t, filepath.Join(dir, fmt.Sprintf("source_%d_test.go", i)), "package syntaxregistration\n"+body)
		} else {
			whole += body
		}
	}
	writeRoleFile39503(t, filepath.Join(dir, "original_test.go"), whole)
	return root, required
}

func writeRoleFile39503(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPhysicalRoleOwnerLayouts39503(t *testing.T) {
	for _, split := range []bool{false, true} {
		root, required := writeRoleLayout39503(t, split)
		owners, err := roleDeclarationOwners39247(root, required)
		if err != nil || len(owners) != 4 {
			t.Fatal("four source declarations lost", split, err)
		}
		seen := map[string]bool{}
		for _, owner := range owners {
			if !seen[owner.logical] {
				verifyRoleOwnerExtraction39247(t, root, owner.logical)
				seen[owner.logical] = true
			}
		}
		if split && len(seen) != 4 || !split && len(seen) != 1 {
			t.Fatal("actual owner layout not retained", seen)
		}
	}
	t.Log("new whole and empty-original/moved-declaration layouts;all4 roots;actual extraction;unchanged75 cap;models0")
}

func TestPhysicalRoleOwnerRefusals39503(t *testing.T) {
	root, required := writeRoleLayout39503(t, true)
	duplicate := filepath.Join(root, "internal/meta/syntaxregistration/duplicate_test.go")
	writeRoleFile39503(t, duplicate, "package syntaxregistration\nfunc RoleAlpha39503() int { return 39617 }\n")
	if _, err := roleDeclarationOwners39247(root, required); err == nil || !strings.Contains(err.Error(), "repeated owners") {
		t.Fatal("duplicate source identity accepted", err)
	}
	if err := os.Remove(duplicate); err != nil {
		t.Fatal(err)
	}
	writeRoleFile39503(t, filepath.Join(root, "internal/meta/syntaxregistration/source_3_test.go"), "package foreign\nfunc RoleDelta39503() int { return 39701 }\n")
	if _, err := roleDeclarationOwners39247(root, required); err == nil || !strings.Contains(err.Error(), "omitted") {
		t.Fatal("missing/foreign owner accepted", err)
	}
	t.Log("new duplicate and foreign/missing ownership inputs rejected;no invented owner")
}
