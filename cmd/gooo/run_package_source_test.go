package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"
)

func TestRunSourceAcceptsPackageDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	directory := filepath.Join("..", "..", "examples", "billing-package")
	handled, code := maybeRunSourcePackage([]string{"--json", "--entry", "PayOrder", directory}, &stdout, &stderr)
	if !handled || code != exitOK {
		t.Fatalf("handled=%t code=%d stderr=%s", handled, code, stderr.String())
	}
	var receipt packageexecution.Receipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || len(receipt.Sources) != 2 {
		t.Fatalf("decision=%s sources=%d", receipt.Decision, len(receipt.Sources))
	}
}

func TestRunSourcePrintsHumanPackageSummary(t *testing.T) {
	var stdout, stderr bytes.Buffer
	directory := filepath.Join("..", "..", "examples", "billing-package")
	code := runSource([]string{"--entry", "PayOrder", directory}, OSFileReader{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	want := "executed package: billing.PayOrder(Order) -> Receipt sources=2 digest="
	if !strings.HasPrefix(stdout.String(), want) {
		t.Fatalf("stdout=%q", stdout.String())
	}
}
