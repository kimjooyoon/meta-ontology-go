package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCheckAndGenerate(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.gooo")
	if err := writeFile(path, `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment`); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check", path}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "ok:") {
		t.Fatalf("check failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	out := filepath.Join(root, "generated")
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"generate", path, "--out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("generate failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	generated, err := os.ReadFile(filepath.Join(out, "semantic.gooo.go"))
	if err != nil || !strings.Contains(string(generated), "func PayOrder") {
		t.Fatalf("generated source = %q, err=%v", generated, err)
	}
}

func TestRunQueryInspectAndAnalyze(t *testing.T) {
	root := t.TempDir()
	goooPath := filepath.Join(root, "main.gooo")
	if err := writeFile(goooPath, `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment`); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"query", "inspect"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{command, goooPath}, &stdout, &stderr); code != exitOK {
			t.Fatalf("%s failed: code=%d out=%q err=%q", command, code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), `"hash"`) || !strings.Contains(stdout.String(), `"nodes"`) {
			t.Fatalf("%s output = %q", command, stdout.String())
		}
	}
	goPath := filepath.Join(root, "semantic.go")
	if err := writeFile(goPath, `package billing

//gooo:semantic activity id="billing://activity/pay" namespace=billing
func Pay(order Order) Payment { return Payment{} }

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct{}

//gooo:semantic entity id="billing://entity/payment" namespace=billing
type Payment struct{}`); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"analyze", goPath}, &stdout, &stderr); code != exitOK {
		t.Fatalf("analyze failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"added"`) || !strings.Contains(stdout.String(), "billing://entity/order") {
		t.Fatalf("analyze output = %q", stdout.String())
	}
}

func TestRunLSPAndExitCodes(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"version"}, &stdout, &stderr); code != exitOK || !strings.Contains(stdout.String(), "gooo dev") {
		t.Fatalf("version failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	if code := run([]string{"unknown"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("unknown command code=%d, want %d", code, exitUsage)
	}
	input := strings.Join([]string{
		lspFrame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`),
		lspFrame(`{"jsonrpc":"2.0","method":"initialized"}`),
		lspFrame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///main.gooo","text":"package billing\nnamespace billing"}}}`),
		lspFrame(`{"jsonrpc":"2.0","id":2,"method":"shutdown"}`),
		lspFrame(`{"jsonrpc":"2.0","method":"exit"}`),
	}, "")
	stdout.Reset()
	stderr.Reset()
	if code := runWithInput([]string{"lsp"}, strings.NewReader(input), &stdout, &stderr); code != exitOK {
		t.Fatalf("lsp failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "textDocument/publishDiagnostics") {
		t.Fatalf("lsp output = %q", stdout.String())
	}
}

func lspFrame(payload string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
