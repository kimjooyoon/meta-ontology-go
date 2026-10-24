package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDensityRewriteRejectsUnavailableModule(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	source := "package p\nfunc Value() int { return 1 }\n"
	for name, content := range map[string]string{
		"go.mod":           "module example.test/root\n",
		"nested/go.mod":    "module example.test/nested\ngo 1.99.0\n",
		"nested/source.go": source,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	_, err := rewriteSubjects(root, []splitSubject{{Logical: "nested/source.go", Lines: densityLines([]byte(source))}})
	if err == nil {
		t.Fatal("density rewrite bypassed the owning module's toolchain")
	}
	got, err := os.ReadFile(filepath.Join(root, "nested", "source.go"))
	if err != nil || string(got) != source {
		t.Fatalf("rejected source changed: %v %s", err, got)
	}
}
