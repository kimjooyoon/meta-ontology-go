package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	runtimeWitnessLineLimit = 75
	runtimeWitnessOutputEnv = "RETURN_TAIL_RUNTIME_WITNESS_OUTPUT_DIR"
	runtimeWitnessGoTimeout = 30 * time.Second
)

type runtimeWitnessCase struct {
	name         string
	functionName string
	source       string
	expected     string
}

func TestReturnTailRuntimeWitness(t *testing.T) {
	cases := []runtimeWitnessCase{
		{
			name:         "W1_sentinel_guard_terminal_nil",
			functionName: "W1",
			source:       runtimeWitnessW1Source(),
			expected: "early:true:*main.witnessError:false\n" +
				"nil:true:<nil>:true\n" +
				"terminal:true:*main.witnessError:false\n",
		},
		{
			name:         "W2_typed_nil_interface",
			functionName: "W2",
			source:       runtimeWitnessW2Source(),
			expected:     "typed-nil:*main.typedNilError:false\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runReturnTailRuntimeWitness(t, tc)
		})
	}
}

func runReturnTailRuntimeWitness(t *testing.T, tc runtimeWitnessCase) {
	t.Helper()
	artifactRoot := os.Getenv(runtimeWitnessOutputEnv)
	if artifactRoot == "" {
		artifactRoot = filepath.Join(t.TempDir(), "return-tail-runtime-witness")
	}
	caseRoot := filepath.Join(artifactRoot, tc.name)
	if err := os.MkdirAll(caseRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	functionLines, renderedInputLines, err := runtimeWitnessInputLines(tc.source, tc.functionName)
	if err != nil {
		t.Fatal(err)
	}
	if functionLines <= runtimeWitnessLineLimit || renderedInputLines <= runtimeWitnessLineLimit {
		t.Fatalf("input capacity was not exceeded: function_lines=%d rendered_input_lines=%d", functionLines, renderedInputLines)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "input-lines.json"), runtimeWitnessJSON(map[string]int{
		"function_lines":        functionLines,
		"rendered_input_lines": renderedInputLines,
	})); err != nil {
		t.Fatal(err)
	}

	originalRoot := filepath.Join(caseRoot, "original")
	if err := runtimeWitnessWriteModule(originalRoot, tc.source); err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "original-source.go"), []byte(tc.source)); err != nil {
		t.Fatal(err)
	}
	originalOutput, originalErr := runtimeWitnessRunGo(originalRoot)
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "original-stdout.txt"), originalOutput); err != nil {
		t.Fatal(err)
	}
	if originalErr != nil {
		t.Fatalf("original fixture did not execute: %v\n%s", originalErr, originalOutput)
	}
	if string(originalOutput) != tc.expected {
		t.Fatalf("original observation=%q, want %q", originalOutput, tc.expected)
	}

	extractionRoot := t.TempDir()
	if err := runtimeWitnessWriteModule(extractionRoot, tc.source); err != nil {
		t.Fatal(err)
	}
	result, extractionErr := ExtractWithResult(extractionRoot, "x.go")
	if extractionErr != nil {
		_ = runtimeWitnessWrite(filepath.Join(caseRoot, "result-error.txt"), []byte(extractionErr.Error()+"\n"))
		t.Fatalf("fixture extraction failed: %v", extractionErr)
	}
	if len(result.Generated) == 0 {
		t.Fatal("fixture extraction produced no generated units")
	}
	evidence, err := json.MarshalIndent(result.Evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "result-evidence.json"), evidence); err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "result-evidence.jsonl"), append(evidence, '\n')); err != nil {
		t.Fatal(err)
	}

	generatedRoot := filepath.Join(caseRoot, "generated")
	generatedArtifactRoot := filepath.Join(caseRoot, "generated-source")
	for logical, data := range result.Generated {
		if err := runtimeWitnessWriteRelative(generatedRoot, logical, data); err != nil {
			t.Fatal(err)
		}
		if err := runtimeWitnessWriteRelative(generatedArtifactRoot, logical, data); err != nil {
			t.Fatal(err)
		}
	}
	unitLines, err := runtimeWitnessGeneratedUnitLines(result.Generated)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "generated-unit-lines.json"), runtimeWitnessJSON(unitLines)); err != nil {
		t.Fatal(err)
	}
	for logical, units := range unitLines {
		for functionName, lineCount := range units {
			if lineCount > runtimeWitnessLineLimit {
				t.Fatalf("generated unit %s:%s has %d raw lines, want <=%d", logical, functionName, lineCount, runtimeWitnessLineLimit)
			}
		}
	}

	generatedOutput, generatedErr := runtimeWitnessRunGo(generatedRoot)
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "generated-stdout.txt"), generatedOutput); err != nil {
		t.Fatal(err)
	}
	if generatedErr != nil {
		t.Fatalf("generated fixture did not execute: %v\n%s", generatedErr, generatedOutput)
	}
	if string(generatedOutput) != tc.expected || string(generatedOutput) != string(originalOutput) {
		t.Fatalf("generated observation=%q, original=%q, want %q", generatedOutput, originalOutput, tc.expected)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "expected-stdout.txt"), []byte(tc.expected)); err != nil {
		t.Fatal(err)
	}
	observations := map[string]string{
		"expected":  tc.expected,
		"original":  string(originalOutput),
		"generated": string(generatedOutput),
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "observations.json"), runtimeWitnessJSON(observations)); err != nil {
		t.Fatal(err)
	}
}

func runtimeWitnessWriteModule(root, source string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := runtimeWitnessWrite(filepath.Join(root, "go.mod"), []byte("module runtime-witness.test\n")); err != nil {
		return err
	}
	return runtimeWitnessWrite(filepath.Join(root, "x.go"), []byte(source))
}

func runtimeWitnessRunGo(root string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runtimeWitnessGoTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", ".")
	command.Dir = root
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=-mod=readonly")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return output, ctx.Err()
	}
	return output, err
}

func runtimeWitnessInputLines(source, functionName string) (int, int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", []byte(source), parser.ParseComments)
	if err != nil {
		return 0, 0, err
	}
	packageLine := fset.Position(file.Pos()).Line
	headerEnd := packageLine
	for _, declaration := range file.Decls {
		if importDeclaration, ok := declaration.(*ast.GenDecl); ok && importDeclaration.Tok.String() == "import" {
			if line := fset.Position(importDeclaration.End()).Line; line > headerEnd {
				headerEnd = line
			}
		}
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name != functionName {
			continue
		}
		start := fset.Position(function.Pos()).Line
		end := fset.Position(function.End()).Line
		functionLines := end - start + 1
		renderedInputLines := end - packageLine + 1
		if headerEnd > start {
			renderedInputLines = headerEnd - packageLine + 1 + functionLines + 1
		}
		return functionLines, renderedInputLines, nil
	}
	return 0, 0, fmt.Errorf("function %q not found", functionName)
}

func runtimeWitnessGeneratedUnitLines(generated map[string][]byte) (map[string]map[string]int, error) {
	result := make(map[string]map[string]int, len(generated))
	for logical, source := range generated {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse generated %s: %w", logical, err)
		}
		packageLine := fset.Position(file.Pos()).Line
		headerEnd := packageLine
		for _, declaration := range file.Decls {
			if importDeclaration, ok := declaration.(*ast.GenDecl); ok && importDeclaration.Tok.String() == "import" {
				if line := fset.Position(importDeclaration.End()).Line; line > headerEnd {
					headerEnd = line
				}
			}
		}
		units := make(map[string]int)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Name == nil {
				continue
			}
			functionLines := fset.Position(function.End()).Line - fset.Position(function.Pos()).Line + 1
			units[function.Name.Name] = headerEnd - packageLine + 1 + functionLines + 1
		}
		result[logical] = units
	}
	return result, nil
}

func runtimeWitnessWriteRelative(root, logical string, data []byte) error {
	clean := filepath.Clean(logical)
	if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("unsafe generated path %q", logical)
	}
	return runtimeWitnessWrite(filepath.Join(root, clean), data)
}

func runtimeWitnessWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func runtimeWitnessJSON(value any) []byte {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return []byte(fmt.Sprintf("{\"serialization_error\":%q}\n", err.Error()))
	}
	return append(data, '\n')
}

func runtimeWitnessW1Source() string {
	return "package main\n\nimport \"fmt\"\n\nfunc W1(mode int) error {\n" +
		strings.Repeat("\t_ = 1\n", 80) +
		"\tif mode == 1 {\n\t\treturn earlySentinel\n\t}\n\tif mode == 2 {\n\t\treturn nil\n\t}\n\treturn terminalSentinel\n}\n\n" +
		"var earlySentinel error = &witnessError{kind: \"early\"}\nvar terminalSentinel error = &witnessError{kind: \"terminal\"}\n\ntype witnessError struct{ kind string }\n\nfunc (e *witnessError) Error() string { return e.kind }\n\nfunc emitW1(label string, got error, expected error) {\n\tfmt.Printf(\"%s:%t:%T:%t\\n\", label, got == expected, got, got == nil)\n}\n\nfunc main() {\n\temitW1(\"early\", W1(1), earlySentinel)\n\temitW1(\"nil\", W1(2), nil)\n\temitW1(\"terminal\", W1(0), terminalSentinel)\n}\n"
}

func runtimeWitnessW2Source() string {
	return "package main\n\nimport \"fmt\"\n\nfunc W2() error {\n" +
		strings.Repeat("\t_ = 1\n", 80) +
		"\treturn (*typedNilError)(nil)\n}\n\ntype typedNilError struct{}\n\nfunc (*typedNilError) Error() string { return \"typed-nil\" }\n\nfunc main() {\n\terr := W2()\n\tfmt.Printf(\"typed-nil:%T:%t\\n\", err, err == nil)\n}\n"
}
