package extractor

import (
	"context"
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

func TestSuffixRejectsRelocatedFunctionLiteralIdentity(t *testing.T) {
	for _, source := range []string{
		"package p; func f() { handler := func() {}; _ = handler }",
		"package p; func f() { invoke(func() {}) }",
		"package p; func f() { for { invoke(func() {}); break } }",
	} {
		file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}
		function := file.Decls[0].(*ast.FuncDecl)
		if !hasUnsafeSuffix(function.Body.List, nil) {
			t.Fatal("suffix accepted a function literal whose enclosing identity changes")
		}
	}
}

func TestSuffixDoesNotRejectAnUnmovedPrefixLiteral(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", "package p; func f() { handler := func() {}; _ = handler; value := 1; _ = value }", 0)
	if err != nil {
		t.Fatal(err)
	}
	function := file.Decls[0].(*ast.FuncDecl)
	if hasUnsafeSuffix(function.Body.List[2:], nil) || hasUnsafeOuterScope(function.Body.List[:2]) {
		t.Fatal("a literal outside the selected suffix became a relocation hazard")
	}
}

func TestSuffixClosureOwnerCounterexampleUsesSameModuleCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("native suffix caller-identity counterexample is CI-only")
	}
	const prefix = "package fixture\nimport (\"runtime\"; \"testing\")\n"
	const body = `t.Run("probe", func(t *testing.T) {
	pc, _, _, ok := runtime.Caller(0)
	if !ok { t.Fatal("caller unavailable") }
	t.Log("OBSERVED_SUFFIX_CALLER=" + runtime.FuncForPC(pc).Name())
})`
	sources := []string{
		prefix + "func TestProbe(t *testing.T) {\n" + body + "\n}\n",
		prefix + "func TestProbe(t *testing.T) { extractedSuffix(t) }\nfunc extractedSuffix(t *testing.T) {\n" + body + "\n}\n",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	callers := make([]string, 0, len(sources))
	for _, source := range sources {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.invalid/suffix-identity\n\ngo 1.27.0\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
		command := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "-count=1", "-v", "-run", "^TestProbe$", ".")
		command.Dir = root
		command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("native identity observation: %v\n%s", err, output)
		}
		observed := ""
		for _, line := range strings.Split(string(output), "\n") {
			if _, value, found := strings.Cut(line, "OBSERVED_SUFFIX_CALLER="); found {
				if observed != "" {
					t.Fatal("duplicate caller observation")
				}
				observed = strings.TrimSpace(value)
			}
		}
		if !strings.HasPrefix(observed, "example.invalid/suffix-identity.") {
			t.Fatalf("caller is not bound to the shared module identity: %q", observed)
		}
		callers = append(callers, observed)
	}
	if callers[0] == callers[1] {
		t.Fatalf("counterexample did not change the callback owner: %q", callers[0])
	}
	t.Logf("same-module suffix counterexample before=%s after=%s", callers[0], callers[1])
}
