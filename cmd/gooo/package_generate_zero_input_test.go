package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestPackageGenerateZeroInputStageKeepsOnlyAuthoredCallerPorts(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "constant.gooo", `package constants
namespace constants
entity Integer id "urn:gooo:type:integer"
activity Seed() -> Integer id "urn:zero:seed" computes "return 7"
activity Main(Integer, Integer) -> Integer id "urn:zero:main" computes "return input0 + input1"
bind Seed.result -> Main.input0
`)
	// Generation stays source-only even for a zero-input exported stage.
	t.Setenv("PATH", t.TempDir())
	for _, row := range []struct {
		entry  string
		inputs int
	}{{"Seed", 0}, {"Main", 1}} {
		manifest := writeWorkspaceManifest(t, root, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"constants","activity":"`+row.entry+`"},"packages":[{"path":"constants","sources":["constant.gooo"]}]}`)
		var out, diagnostics bytes.Buffer
		if code := run([]string{"package", "generate", "--go-package", "constants",
			"--export-stage", "Constant=urn:zero:seed", manifest}, &out, &diagnostics); code != exitOK || diagnostics.Len() != 0 {
			t.Fatal(row.entry, code, diagnostics.String())
		}
		file, err := parser.ParseFile(token.NewFileSet(), "generated.go", out.Bytes(), 0)
		if err != nil {
			t.Fatal(err)
		}
		functions := map[string]int{}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok {
				functions[function.Name.Name] = function.Type.Params.NumFields()
			}
		}
		constantInputs, constant := functions["Constant"]
		callerInputs, entry := functions["Evaluate"]
		if !constant || constantInputs != 0 || !entry || callerInputs != row.inputs {
			t.Fatal("generated API invented caller inputs", row.entry, functions)
		}
	}
}
