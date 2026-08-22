package feedbackstate

import (
	"bytes"
	"go/format"
	"os"
	"testing"
)

func TestCIFormatReceipt(t *testing.T) {
	names := []string{
		"contract.go", "evaluate.go", "evaluate_test.go", "evidence.go",
		"fixture_test.go", "format_receipt_test.go", "receipt.go", "semantics.go",
	}
	different := false
	for _, name := range names {
		source, err := os.ReadFile(name)
		if err != nil { t.Fatal(err) }
		canonical, err := format.Source(source)
		if err != nil { t.Fatalf("%s: %v", name, err) }
		if bytes.Equal(source, canonical) { continue }
		different = true
		t.Logf("=== %s ===\n%s=== END %s ===", name, canonical, name)
	}
	if different { t.Fatal("CI_FORMAT_RECEIPT") }
}
