package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
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
	receipt := decodePackageReceipt(t, stdout.Bytes())
	if receipt.Decision != "PASS" || len(receipt.Sources) != 2 {
		t.Fatalf("decision=%s sources=%d", receipt.Decision, len(receipt.Sources))
	}
}

func TestRunSourceJSONPackageReplayIsByteStable(t *testing.T) {
	directory := filepath.Join("..", "..", "examples", "billing-package")
	var firstOut, firstErr bytes.Buffer
	var secondOut, secondErr bytes.Buffer
	firstCode := runSource([]string{"--json", "--entry", "PayOrder", directory}, OSFileReader{}, &firstOut, &firstErr)
	secondCode := runSource([]string{"--json", "--entry", "PayOrder", directory}, OSFileReader{}, &secondOut, &secondErr)
	if firstCode != exitOK || secondCode != exitOK || firstErr.Len() != 0 || secondErr.Len() != 0 || !bytes.Equal(firstOut.Bytes(), secondOut.Bytes()) {
		t.Fatalf("package JSON replay was not stable: first=(%d,%q,%q), second=(%d,%q,%q)", firstCode, firstOut.String(), firstErr.String(), secondCode, secondOut.String(), secondErr.String())
	}
	receipt := decodePackageReceipt(t, firstOut.Bytes())
	if receipt.Digest == "" || receipt.Decision != "PASS" {
		t.Fatalf("unexpected replay receipt: decision=%s digest=%q", receipt.Decision, receipt.Digest)
	}
}

func TestRunSourceJSONPackageBoundaryFailuresAreSealed(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "only.gooo"), []byte("package billing\nnamespace billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "PayOrder", directory}, OSFileReader{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("JSON package boundary failure = %d, stderr=%q", code, stderr.String())
	}
	receipt := decodePackageReceipt(t, stdout.Bytes())
	if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "PACKAGE_SOURCE_DIRECTORY_UNAVAILABLE" || len(receipt.Diagnostics) != 1 {
		t.Fatalf("unexpected boundary receipt: %#v", receipt)
	}
}

func TestRunSourceJSONUnsupportedPackageInvocationIsSealed(t *testing.T) {
	directory := filepath.Join("..", "..", "examples", "billing-package")
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "PayOrder", directory, "extra"}, OSFileReader{}, &stdout, &stderr)
	if code != exitUsage || stderr.Len() != 0 {
		t.Fatalf("JSON unsupported invocation = %d, stderr=%q", code, stderr.String())
	}
	receipt := decodePackageReceipt(t, stdout.Bytes())
	if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "PACKAGE_INVOCATION_UNSUPPORTED" || len(receipt.Diagnostics) != 1 {
		t.Fatalf("unexpected unsupported invocation receipt: %#v", receipt)
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

func decodePackageReceipt(t *testing.T, data []byte) packageexecution.Receipt {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var receipt packageexecution.Receipt
	if err := decoder.Decode(&receipt); err != nil {
		t.Fatalf("package JSON did not decode: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("package JSON had trailing data: %v", err)
	}
	if err := packageexecution.Validate(receipt); err != nil {
		t.Fatalf("package JSON receipt was not sealed: %v", err)
	}
	return receipt
}
