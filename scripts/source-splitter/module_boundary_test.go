package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitPlanRespectsModuleBoundary(t *testing.T) {
	for _, test := range []struct {
		name, version, wantError string
		crossModule              bool
	}{
		{name: "nested source", version: "1.27.3"},
		{name: "newer toolchain", version: "1.99.0", wantError: "toolchain"},
		{name: "source link", version: "1.27.3", crossModule: true, wantError: "module boundary"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "nested"), 0o700); err != nil {
				t.Fatal(err)
			}
			source := "package p\n\nfunc First() int { return 1 }\n\nfunc Second() int { return 2 }\n"
			files := map[string]string{
				"go.mod":        "module example.test/root\n",
				"nested/go.mod": "module example.test/nested\n\ngo " + test.version + "\n",
				"root.go":       source,
			}
			if !test.crossModule {
				files["nested/source.go"] = source
			}
			for name, content := range files {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if test.crossModule {
				if err := os.Symlink("../root.go", filepath.Join(root, "nested", "source.go")); err != nil {
					t.Fatal(err)
				}
			}
			plan, err := planSource(root, "nested/source.go", 4)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) || len(plan.Parts) != 0 {
					t.Fatalf("unsafe plan accepted: parts=%d err=%v", len(plan.Parts), err)
				}
				return
			}
			if err != nil || len(plan.Parts) != 2 {
				t.Fatalf("owning module plan failed: parts=%d err=%v", len(plan.Parts), err)
			}
			for _, part := range plan.Parts {
				if !strings.HasPrefix(part.Subject, "nested/") {
					t.Fatalf("split escaped owning module: %s", part.Subject)
				}
			}
		})
	}
}
