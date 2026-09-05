package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitPartPathPreservesBuildDomain(t *testing.T) {
	tests := map[string]string{
		"fixture.go":                  "fixture_split02.go",
		"fixture_test.go":             "fixture_split02_test.go",
		"fixture_linux.go":            "fixture_split02_linux.go",
		"fixture_linux_amd64_test.go": "fixture_split02_linux_amd64_test.go",
	}
	for input, expected := range tests {
		actual, err := splitPartPath(input, 2)
		if err != nil || actual != expected {
			t.Fatalf("splitPartPath(%q) = %q, %v; want %q", input, actual, err, expected)
		}
	}
}

func TestPlanSourceSplitsDeclarationsAndImports(t *testing.T) {
	root := t.TempDir()
	subject := "fixture_linux_test.go"
	source := `//go:build linux

package fixture

import "fmt"

func first() string {
	return fmt.Sprint("a")
}

func second() int {
	return 2
}
`
	if err := os.WriteFile(filepath.Join(root, subject), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := planSource(root, subject, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parts) != 2 || !strings.HasSuffix(plan.Parts[1].Subject, "_split02_linux_test.go") {
		t.Fatalf("unexpected parts: %#v", plan.Parts)
	}
	for _, part := range plan.Parts {
		if physicalLines(part.Data) > 10 || !strings.Contains(string(part.Data), "//go:build linux") {
			t.Fatalf("invalid split part %s:\n%s", part.Subject, part.Data)
		}
	}
	for _, part := range plan.Parts {
		data := string(part.Data)
		if strings.Contains(data, `"fmt"`) != strings.Contains(data, "fmt.Sprint") {
			t.Fatalf("import usage mismatch:\n%s", part.Data)
		}
	}
}

func TestPlanSourceFunctionExtractorApplyUsesSharedImportHeader(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "function-extractor-apply.go.txt"))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	subject := "apply.go"
	if err := os.WriteFile(filepath.Join(root, subject), source, 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := planSource(root, subject, 75)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, part := range plan.Parts {
		text := string(part.Data)
		if !strings.Contains(text, "func stageExtractions(") {
			continue
		}
		if found {
			t.Fatal("stageExtractions was rendered into multiple split parts")
		}
		found = true
		if physicalLines(part.Data) > 75 || strings.Contains(text, "import (\n") {
			t.Fatalf("stageExtractions helper exceeded rendered capacity or retained grouped imports:\n%s", text)
		}
		imports := []string{
			`import "bytes"`,
			`import "fmt"`,
			`import "os"`,
			`import recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"`,
			`import projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"`,
		}
		previous := -1
		for _, imported := range imports {
			position := strings.Index(text, imported)
			if position < 0 || position <= previous {
				t.Fatalf("shared import header lost alias/path/order %q: %q", imported, text)
			}
			previous = position
		}
		for _, body := range []string{
			"operations, evidence, err := stageGenericExtraction(root, logical, buffers, created, changedBySubject, createdBySubject)",
			"changedBySubject[logical] = appendUnique(changedBySubject[logical], edit.Path)",
			"return staged, subjects, unhandled, failures, nil",
		} {
			if !strings.Contains(text, body) {
				t.Fatalf("shared render lost stageExtractions body %q: %q", body, text)
			}
		}
	}
	if !found {
		t.Fatal("stageExtractions was not assigned to a split part")
	}
}

func TestRenderPartSharedImportHeaderGuards(t *testing.T) {
	cases := []struct {
		name       string
		source     string
		grouped    bool
		wantText   string
	}{
		{name: "named-alias", source: "package p\n\nimport (\n\tj \"encoding/json\"\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ j.RawMessage\n\t_ = strconv.IntSize\n}\n", grouped: false, wantText: `import j "encoding/json"`},
		{name: "dot-import-preserved", source: "package p\n\nimport (\n\t. \"encoding/json\"\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ RawMessage\n\t_ = strconv.IntSize\n}\n", grouped: true, wantText: `. "encoding/json"`},
		{name: "blank-import-preserved", source: "package p\n\nimport (\n\t_ \"encoding/json\"\n\t\"strconv\"\n)\n\nfunc F() {\n\t_ = strconv.IntSize\n}\n", grouped: true, wantText: `_ "encoding/json"`},
		{name: "group-comment-preserved", source: "package p\n\n// retain import documentation\nimport (\n\t\"encoding/json\"\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ json.RawMessage\n\t_ = strconv.IntSize\n}\n", grouped: true, wantText: "retain import documentation"},
		{name: "detached-comment-preserved", source: "package p\n\nimport (\n\t\"encoding/json\"\n\n\t// detached import note\n\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ json.RawMessage\n\t_ = strconv.IntSize\n}\n", grouped: true, wantText: "detached import note"},
		{name: "cgo-import-preserved", source: "package p\n\nimport (\n\t\"C\"\n\t\"strconv\"\n)\n\nfunc F() {\n\t_ = C.symbol\n\t_ = strconv.IntSize\n}\n", grouped: true, wantText: `"C"`},
	}
	if len(cases) != 6 {
		t.Fatalf("shared import header guard cohort denominator=%d, want 6", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", []byte(tc.source), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var function *ast.FuncDecl
			for _, declaration := range file.Decls {
				candidate, ok := declaration.(*ast.FuncDecl)
				if ok && candidate.Name != nil && candidate.Name.Name == "F" {
					function = candidate
				}
			}
			if function == nil {
				t.Fatal("function fixture missing")
			}
			data, err := renderPart(fset, file, []ast.Decl{function})
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if strings.Contains(text, "import (\n") != tc.grouped {
				t.Fatalf("grouped header=%t, want %t: %q", strings.Contains(text, "import (\n"), tc.grouped, text)
			}
			if !strings.Contains(text, tc.wantText) {
				t.Fatalf("header lost guard text %q: %q", tc.wantText, text)
			}
		})
	}
}
