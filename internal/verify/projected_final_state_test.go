package verify

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestProjectedCompileObservesFinalTransformation(t *testing.T) {
	for _, closure := range []string{"partial", "closed"} {
		for _, value := range []string{"0", "1"} {
			t.Run(closure+"/"+value, func(t *testing.T) {
				checkFinalProjection(t, closure, value)
			})
		}
	}
}

func checkFinalProjection(t *testing.T, closure, value string) {
	root, command := finalProjectionFixture(t, closure, value)
	output, err := command.CombinedOutput()
	wantPass := value == "0"
	if (err == nil) != wantPass {
		t.Fatalf("final source outcome lost: %v\n%s", err, output)
	}
	if got := string(finalProjectionRead(t, root, "events")); got != "rewrite\ncompile\n" {
		t.Fatalf("expected one verification after the last write; got %q", got)
	}
	name := "projected-compile.json"
	if closure == "partial" {
		name = "fixed-point-unknown.json"
	}
	if !wantPass {
		name = "projected-compile-failure.json"
	}
	var receipt struct {
		Compile struct {
			Observed bool `json:"observed"`
			ExitCode int  `json:"exit_code"`
		} `json:"projected_compile"`
	}
	raw := finalProjectionRead(t, root, "projection/"+name)
	if name == "projected-compile.json" {
		raw = append(append([]byte(`{"projected_compile":`), raw...), '}')
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Compile.Observed != wantPass || (receipt.Compile.ExitCode == 0) != wantPass {
		t.Fatal("receipt carried another program state's result", receipt)
	}
}

func finalProjectionFixture(t *testing.T, closure, value string) (string, *exec.Cmd) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"go.mod":         "module sample\n\ngo " + strings.TrimPrefix(runtime.Version(), "go") + "\n",
		"sample.go":      "package sample\nfunc Value() int { return 0 }\n",
		"sample_test.go": "package sample\nimport \"testing\"\nfunc TestValue(t *testing.T) { if Value() != 0 { t.Fatal(\"source behavior changed after validation\") } }\n",
		"harness.sh":     finalProjectionHarness(t),
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, "bash", filepath.Join(root, "harness.sh"), root)
	command.Env = append(os.Environ(), "REAL_GO="+filepath.Join(runtime.GOROOT(), "bin", "go"),
		"VALUE="+value, "CLOSURE="+closure, "GOWORK=off", "GOTOOLCHAIN=local")
	return root, command
}

func finalProjectionRead(t *testing.T, root, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
