package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageexecution"
)

func TestRunSourceAcceptsPackageDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	directory := filepath.Join("..", "..", "examples", "billing-package")
	handled, code := maybeRunSourcePackage([]string{"--entry", "PayOrder", directory}, &stdout, &stderr)
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
