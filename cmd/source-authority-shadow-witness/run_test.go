package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, root string) (string, string) {
	t.Helper()
	raw, err := os.ReadFile("../../examples/language-assurance-kernel/source-authority-shadow.json")
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.Repeat("a", 40)
	raw = bytes.Replace(raw, []byte(strings.Repeat("0", 40)), []byte(sha), 1)
	input := filepath.Join(root, "input.json")
	if err := os.WriteFile(input, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return input, sha
}

func TestRunWritesSatisfiedShadowReceiptOutsideRoot(t *testing.T) {
	root := t.TempDir()
	input, sha := writeFixture(t, root)
	output := filepath.Join(t.TempDir(), "receipt.json")
	failClosed, err := run(config{root: root, input: input, output: output,
		expectedSHA: sha, check: true})
	if err != nil || failClosed {
		t.Fatalf("run failClosed=%v err=%v", failClosed, err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("receipt state info=%v err=%v", info, err)
	}
}

func TestRunRejectsRepositoryOutput(t *testing.T) {
	root := t.TempDir()
	input, sha := writeFixture(t, root)
	output := filepath.Join(root, "receipt.json")
	if _, err := run(config{root: root, input: input, output: output,
		expectedSHA: sha}); err == nil {
		t.Fatal("repository output was accepted")
	}
}

func TestRunDoesNotOverwriteReceipt(t *testing.T) {
	root := t.TempDir()
	input, sha := writeFixture(t, root)
	output := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(output, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(config{root: root, input: input, output: output,
		expectedSHA: sha}); err == nil {
		t.Fatal("existing receipt was overwritten")
	}
}
