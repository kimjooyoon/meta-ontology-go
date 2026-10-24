package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func TestCurrentRecipeStagingReachesDeclarationCapacity(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	subject := "bootstrap/function-extractor/apply.go"
	source, err := os.ReadFile(filepath.Join(root, subject))
	if err != nil {
		t.Fatal(err)
	}
	// Repository projection also runs this test after the authorized split.
	if physicalLines(source) <= 75 {
		return
	}
	plan, err := planSource(root, subject, 75)
	if err != nil {
		t.Fatalf("current recipe staging cannot reach the split consumer: %v", err)
	}
	for _, part := range plan.Parts {
		if physicalLines(part.Data) > 75 {
			t.Fatalf("part exceeds the consumer's declaration capacity: %s", part.Subject)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, plan.Directory))
	if err != nil {
		t.Fatal(err)
	}
	directory := metricevidence.Directory{Path: plan.Directory}
	for _, entry := range entries {
		if entry.IsDir() {
			directory.DirectFolders++
		} else {
			directory.DirectFiles++
		}
	}
	report := metricevidence.Report{Directories: []metricevidence.Directory{directory}}
	report.Meta.Policy.MaxDirectEntries = 10
	if err := validateTopology(report, plan); err != nil {
		t.Fatalf("current recipe staging exceeds the split consumer's directory capacity: %v", err)
	}
}

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
