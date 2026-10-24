package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecipeWriteSetStaysInOwningModule(t *testing.T) {
	for _, kind := range []string{"nested", "cross-edit", "cross-creation", "newer", "missing-import"} {
		t.Run(kind, func(t *testing.T) {
			t.Setenv("GOPROXY", "off")
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "nested"), 0o700); err != nil {
				t.Fatal(err)
			}
			version := "1.27.3"
			if kind == "newer" {
				version = "1.99.0"
			}
			for name, content := range map[string]string{
				"go.mod": "module example.test/root\n", "root.go": "package p\nfunc Root() int { return 1 }\n",
				"nested/go.mod":    "module example.test/nested\ngo " + version + "\n",
				"nested/source.go": "package p\nfunc Value() int { return 1 }\n",
			} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			recipe := extractionRecipe{Subject: "nested/source.go", Operation: "extract-function",
				Edits:   []textEdit{{Path: "nested/source.go", Old: []string{"return 1"}, New: []string{"return 2"}}},
				Creates: []fileCreation{{Path: "nested/helper.go", Lines: []string{"package p", "func Helper() int { return 2 }"}}}}
			if kind == "cross-edit" {
				recipe.Edits = append(recipe.Edits, textEdit{Path: "root.go", Old: []string{"return 1"}, New: []string{"return 2"}})
			}
			if kind == "cross-creation" {
				recipe.Creates[0].Path = "helper.go"
			}
			if kind == "missing-import" {
				recipe.Creates[0].Lines = []string{"package p", `import "example.invalid/not-declared"`,
					"func Helper() int { return missing.Value() }"}
			}
			staged, _, _, _, err := stageExtractions(root, map[string]planSubject{recipe.Subject: {}},
				[]string{recipe.Subject}, []extractionRecipe{recipe})
			if kind == "nested" {
				if err != nil || len(staged) != 2 || staged["nested/helper.go"].data == nil {
					t.Fatalf("valid nested recipe did not stage: %v %v", staged, err)
				}
			} else if err == nil || len(staged) != 0 {
				t.Fatalf("recipe bypassed its owning module: %v %v", staged, err)
			}
		})
	}
}
