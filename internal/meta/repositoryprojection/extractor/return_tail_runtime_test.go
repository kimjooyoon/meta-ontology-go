package extractor

import (
	"bytes"
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
	support      map[string]string
	expected     string
}

func TestReturnTailRuntimeWitness(t *testing.T) {
	cases := []runtimeWitnessCase{
		{
			name:         "W1_sentinel_guard_terminal_nil",
			functionName: "W1",
			source:       runtimeWitnessW1Source(),
			support:      map[string]string{"harness.go": runtimeWitnessW1Harness()},
			expected: "early:true:*main.witnessError:false\n" +
				"nil:true:<nil>:true\n" +
				"terminal:true:*main.witnessError:false\n",
		},
		{
			name:         "W2_typed_nil_interface",
			functionName: "W2",
			source:       runtimeWitnessW2Source(),
			support:      map[string]string{"harness.go": runtimeWitnessW2Harness()},
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
	if os.Getenv("CI") != "true" {
		t.Skip("runtime witness requires CI=true")
	}
	caseRoot := runtimeWitnessPrepareCase(t, tc)
	originalOutput := runtimeWitnessRunOriginal(t, tc, caseRoot)
	result := runtimeWitnessExtract(t, tc, caseRoot)
	generatedRoot := runtimeWitnessMaterializeGenerated(t, tc, result, caseRoot)
	generatedOutput := runtimeWitnessRunGenerated(t, tc, generatedRoot, caseRoot)
	runtimeWitnessAssertOutputs(t, tc, originalOutput, generatedOutput, caseRoot)
}

func runtimeWitnessPrepareCase(t *testing.T, tc runtimeWitnessCase) string {
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
	metadata := map[string]any{"function": tc.functionName, "expected_output": tc.expected, "function_lines": functionLines, "rendered_input_lines": renderedInputLines}
	for path, data := range map[string][]byte{
		"source.go":           []byte(tc.source),
		"expected-output.txt": []byte(tc.expected),
		"case-metadata.json":  runtimeWitnessJSON(metadata),
	} {
		if err := runtimeWitnessWrite(filepath.Join(caseRoot, path), data); err != nil {
			t.Fatal(err)
		}
	}
	for logical, source := range tc.support {
		if err := runtimeWitnessWriteRelative(filepath.Join(caseRoot, "support-source"), logical, []byte(source)); err != nil {
			t.Fatal(err)
		}
	}
	return caseRoot
}

func runtimeWitnessRunOriginal(t *testing.T, tc runtimeWitnessCase, caseRoot string) []byte {
	t.Helper()
	root := t.TempDir()
	if err := runtimeWitnessWriteModule(root, tc.source, tc.support); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, runErr := runtimeWitnessRunGo(root)
	for path, data := range map[string][]byte{
		"original-stdout.txt": []byte(stdout),
		"original-stderr.txt": []byte(stderr),
	} {
		if err := runtimeWitnessWrite(filepath.Join(caseRoot, path), data); err != nil {
			t.Fatal(err)
		}
	}
	if runErr != nil {
		t.Fatalf("original fixture did not execute: %v\n%s", runErr, stderr)
	}
	if string(stdout) != tc.expected {
		t.Fatalf("original observation=%q, want %q", stdout, tc.expected)
	}
	return stdout
}

func runtimeWitnessExtract(t *testing.T, tc runtimeWitnessCase, caseRoot string) Result {
	t.Helper()
	root := t.TempDir()
	if err := runtimeWitnessWriteModule(root, tc.source, tc.support); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, "x.go")
	if err != nil {
		if writeErr := runtimeWitnessWrite(filepath.Join(caseRoot, "result-error.txt"), []byte(err.Error()+"\n")); writeErr != nil {
			t.Fatal(writeErr)
		}
		t.Fatalf("fixture extraction failed: %v", err)
	}
	if len(result.Generated) == 0 {
		t.Fatal("fixture extraction produced no generated units")
	}
	if tc.name == "W1_sentinel_guard_terminal_nil" {
		assertRuntimeWitnessW1DependencyEvidence(t, result)
	}
	evidence, err := json.MarshalIndent(result.Evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "result-evidence.json"), evidence); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertRuntimeWitnessW1DependencyEvidence(t *testing.T, result Result) {
	t.Helper()
	witnesses := make([]StrategyEvidence, 0)
	for _, evidence := range result.Evidence {
		if evidence.Subject == "func:W1" {
			witnesses = append(witnesses, evidence)
		}
	}
	if len(witnesses) < 2 {
		t.Fatalf("W1 runtime evidence=%+v, want multiple prepare stages", witnesses)
	}
	linkedDependencies := 0
	for index, current := range witnesses {
		for _, dependency := range current.CalleeDependencies {
			linked := false
			for previousIndex := 0; previousIndex < index; previousIndex++ {
				previous := witnesses[previousIndex]
				if previous.Helper == dependency.Name && len(previous.ProofStages) > 3 && dependency.EvidenceID != "" && dependency.EvidenceID == previous.ProofStages[3].OutputEvidenceID {
					linked = true
					break
				}
			}
			if !linked {
				t.Fatalf("W1 evidence[%d] dependency=%+v lacks an earlier matching helper callee-effects proof", index, dependency)
			}
			linkedDependencies++
			if !runtimeWitnessGeneratedFunctionExists(result.Generated, dependency.Name) {
				t.Fatalf("W1 dependency helper %q was not present in generated output", dependency.Name)
			}
		}
	}
	if linkedDependencies == 0 {
		t.Fatalf("W1 runtime evidence=%+v, want at least one generated dependency link", witnesses)
	}
}

func runtimeWitnessGeneratedFunctionExists(generated map[string][]byte, name string) bool {
	needle := "func " + name + "("
	for _, source := range generated {
		if strings.Contains(string(source), needle) {
			return true
		}
	}
	return false
}

func runtimeWitnessMaterializeGenerated(t *testing.T, tc runtimeWitnessCase, result Result, caseRoot string) string {
	t.Helper()
	root := t.TempDir()
	if err := runtimeWitnessWriteGoMod(root); err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWriteSupport(root, tc.support); err != nil {
		t.Fatal(err)
	}
	generatedArtifactRoot := filepath.Join(caseRoot, "generated-source")
	for logical, data := range result.Generated {
		if err := runtimeWitnessWriteRelative(root, logical, data); err != nil {
			t.Fatal(err)
		}
		if err := runtimeWitnessWriteRelative(generatedArtifactRoot, logical, data); err != nil {
			t.Fatal(err)
		}
	}
	fileLines := runtimeWitnessGeneratedFileLines(result.Generated)
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "generated-file-lines.json"), runtimeWitnessJSON(fileLines)); err != nil {
		t.Fatal(err)
	}
	for logical, lineCount := range fileLines {
		if lineCount > runtimeWitnessLineLimit {
			t.Fatalf("generated file %s has %d raw lines, want <=%d", logical, lineCount, runtimeWitnessLineLimit)
		}
	}
	unitLines, err := runtimeWitnessGeneratedFunctionUnitLines(result.Generated)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "generated-function-unit-lines.json"), runtimeWitnessJSON(unitLines)); err != nil {
		t.Fatal(err)
	}
	for logical, units := range unitLines {
		for functionName, lineCount := range units {
			if lineCount > runtimeWitnessLineLimit {
				t.Fatalf("generated function unit %s:%s has %d raw lines, want <=%d", logical, functionName, lineCount, runtimeWitnessLineLimit)
			}
		}
	}
	return root
}

func runtimeWitnessRunGenerated(t *testing.T, tc runtimeWitnessCase, root, caseRoot string) []byte {
	t.Helper()
	stdout, stderr, runErr := runtimeWitnessRunGo(root)
	for path, data := range map[string][]byte{
		"generated-stdout.txt": []byte(stdout),
		"generated-stderr.txt": []byte(stderr),
	} {
		if err := runtimeWitnessWrite(filepath.Join(caseRoot, path), data); err != nil {
			t.Fatal(err)
		}
	}
	if runErr != nil {
		t.Fatalf("generated fixture did not execute: %v\n%s", runErr, stderr)
	}
	return stdout
}

func runtimeWitnessAssertOutputs(t *testing.T, tc runtimeWitnessCase, original, generated []byte, caseRoot string) {
	t.Helper()
	if string(generated) != tc.expected || string(generated) != string(original) {
		t.Fatalf("generated observation=%q, original=%q, want %q", generated, original, tc.expected)
	}
	observations := map[string]string{
		"expected_stdout":  tc.expected,
		"original_stdout":  string(original),
		"generated_stdout": string(generated),
	}
	if err := runtimeWitnessWrite(filepath.Join(caseRoot, "observations.json"), runtimeWitnessJSON(observations)); err != nil {
		t.Fatal(err)
	}
}

func runtimeWitnessWriteModule(root, source string, support map[string]string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := runtimeWitnessWriteGoMod(root); err != nil {
		return err
	}
	if err := runtimeWitnessWrite(filepath.Join(root, "x.go"), []byte(source)); err != nil {
		return err
	}
	return runtimeWitnessWriteSupport(root, support)
}

func runtimeWitnessWriteSupport(root string, support map[string]string) error {
	for logical, source := range support {
		if err := runtimeWitnessWriteRelative(root, logical, []byte(source)); err != nil {
			return err
		}
	}
	return nil
}

func runtimeWitnessWriteGoMod(root string) error {
	return runtimeWitnessWrite(filepath.Join(root, "go.mod"), []byte("module runtime-witness.test\n"))
}

func runtimeWitnessRunGo(root string) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runtimeWitnessGoTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", ".")
	command.Dir = root
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=-mod=readonly")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	if ctx.Err() != nil {
		return stdout, stderr.Bytes(), ctx.Err()
	}
	return stdout, stderr.Bytes(), err
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

func runtimeWitnessGeneratedFileLines(generated map[string][]byte) map[string]int {
	result := make(map[string]int, len(generated))
	for logical, source := range generated {
		result[logical] = runtimeWitnessRawPhysicalLines(source)
	}
	return result
}

func runtimeWitnessRawPhysicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func runtimeWitnessGeneratedFunctionUnitLines(generated map[string][]byte) (map[string]map[string]int, error) {
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
	return "package main\n\nfunc W1(mode int) error {\n" +
		"\tif mode == 1 {\n\t\treturn earlySentinel\n\t}\n" +
		strings.Repeat("\t_ = 1\n", 80) +
		"\tif mode == 2 {\n\t\treturn nil\n\t}\n\treturn terminalSentinel\n}\n"
}

func runtimeWitnessW1Harness() string {
	return "package main\n\nimport \"fmt\"\n\nvar earlySentinel error = &witnessError{kind: \"early\"}\nvar terminalSentinel error = &witnessError{kind: \"terminal\"}\n\ntype witnessError struct{ kind string }\n\nfunc (e *witnessError) Error() string { return e.kind }\n\nfunc emitW1(label string, got error, expected error) {\n\tfmt.Printf(\"%s:%t:%T:%t\\n\", label, got == expected, got, got == nil)\n}\n\nfunc main() {\n\temitW1(\"early\", W1(1), earlySentinel)\n\temitW1(\"nil\", W1(2), nil)\n\temitW1(\"terminal\", W1(0), terminalSentinel)\n}\n"
}

func runtimeWitnessW2Source() string {
	return "package main\n\nfunc W2() error {\n" +
		strings.Repeat("\t_ = 1\n", 80) +
		"\treturn (*typedNilError)(nil)\n}\n"
}

func runtimeWitnessW2Harness() string {
	return "package main\n\nimport \"fmt\"\n\ntype typedNilError struct{}\n\nfunc (*typedNilError) Error() string { return \"typed-nil\" }\n\nfunc main() {\n\terr := W2()\n\tfmt.Printf(\"typed-nil:%T:%t\\n\", err, err == nil)\n}\n"
}
