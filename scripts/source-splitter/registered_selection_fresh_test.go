package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisteredSelectionSplit11447(t *testing.T) {
	root := os.Getenv("GOOO_REGISTERED_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	maximum, functions, parts := 0, 0, 0
	for name, raw := range selectionCapacitySources11447(t, root) {
		limit, count, output := checkSelectionUnit11447(t, name, raw)
		maximum, functions, parts = max(maximum, limit), functions+count, parts+output
	}
	if functions < 6 || maximum > 45 {
		t.Fatal("selection chunks were not bounded", functions, maximum)
	}
	if dir := os.Getenv("GOOO_SELECTION_OUTPUT"); dir != "" {
		value := map[string]any{"selection_functions": functions, "maximum_function_lines": maximum,
			"split_parts": parts, "line_limit": 75, "language_program_executions": 0}
		encoded, _ := json.MarshalIndent(value, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "selection11447.json"), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("registered selector functions=%d maximum=%d lines; actual splitter parts=%d each<=75; language executions0", functions, maximum, parts)
}

func checkSelectionUnit11447(t *testing.T, name string, raw []byte) (int, int, int) {
	t.Helper()
	temporary := t.TempDir()
	if err := os.WriteFile(filepath.Join(temporary, name), raw, 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := planSource(temporary, name, 75)
	if err != nil {
		t.Fatal("registered selection cannot preserve source order", err)
	}
	maximum, functions := 0, 0
	for _, part := range plan.Parts {
		if physicalLines(part.Data) > 75 {
			t.Fatal("split part exceeds 75 lines", part.Subject)
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, part.Subject, part.Data, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(function.Name.Name, "activeDenominator") {
				continue
			}
			lines := set.Position(function.End()).Line - set.Position(function.Pos()).Line + 1
			maximum = max(maximum, lines)
			functions++
		}
	}

	return maximum, functions, len(plan.Parts)
}

func selectionCapacitySources11447(t *testing.T, root string) map[string][]byte {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, "internal/meta/languageassurance/verticalsliceclosureshadow", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, raw, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(function.Name.Name, "activeDenominator") {
				out[filepath.Base(path)] = raw
				break
			}
		}
	}
	return out
}
