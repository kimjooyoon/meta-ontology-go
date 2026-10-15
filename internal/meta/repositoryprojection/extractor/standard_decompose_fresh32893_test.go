package extractor

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSourceOwnedStandardDecomposition32893(t *testing.T) {
	root := freshStandardRoot31711(t)
	source := "package main\nimport (\"context\";\"fmt\")\nfunc F(ctx context.Context, x int) int {\n" +
		strings.Repeat("x += 1\n", 101) + "return x\n}\nfunc main(){for _, x:=range []int{-4,0,8}{fmt.Println(F(context.Background(),x))}}\n"
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	original := freshRunSource32893(t, root)
	unsupported, err := ExtractWithResult(root, "main.go")
	if err == nil || len(unsupported.Generated) != 0 || !strings.Contains(err.Error(), "NO_SAFE_DECLARATION_CAPACITY") {
		t.Fatal("compound-update source acquired an unproven rewrite", err)
	}
	source = "package main\nimport (\"context\";\"fmt\")\nfunc F(ctx context.Context, x int) int {\n" +
		strings.Repeat("_ = ctx\n", 101) + "return x + 101\n}\nfunc main(){for _, x:=range []int{-4,0,8}{fmt.Println(F(context.Background(),x))}}\n"
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, "main.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Evidence) == 0 || len(result.Generated) < 2 {
		t.Fatal("source-bound decomposition did not generate helpers")
	}
	for name, raw := range result.Generated {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	generated := freshRunSource32893(t, root)
	if !bytes.Equal(original, generated) || string(generated) != "97\n101\n109\n" {
		t.Fatalf("generated source changed behavior: %q / %q", original, generated)
	}
	t.Log("one new source:compound101updates remain unproven/no rewrite;distinct context-typed101reads+addition source creates Gooo-owned helpers;3 runtime inputs preserve exact results;model0")
}

func freshRunSource32893(t *testing.T, root string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	goPath := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goPath += ".exe"
	}
	command := exec.CommandContext(ctx, goPath, "run", "-mod=readonly", ".")
	command.Dir = root
	command.Env = moduleListEnvironment(os.Environ())
	raw, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("new source execution failed: %v / %s", err, raw)
	}
	return raw
}
