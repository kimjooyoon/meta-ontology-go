package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReceiptSchemaCommandUsesActualGoooSource(t *testing.T) {
	for _, format := range []string{"go", "json"} {
		var out, errOut bytes.Buffer
		code := run([]string{"receipt-schema", "--format", format, "../../internal/completeness/receipt.gooo"}, &out, &errOut)
		if code != exitOK {
			t.Fatalf("exit=%d %s", code, errOut.String())
		}
		if format == "json" && !json.Valid(out.Bytes()) {
			t.Fatal("invalid JSON schema")
		}
		if format == "go" && !strings.Contains(out.String(), "type CompletenessReceipt struct") {
			t.Fatal("missing generated structure")
		}
	}
}

func TestReceiptSchemaCommandRejectsMissingAndUnknownInputs(t *testing.T) {
	for _, args := range [][]string{{"receipt-schema"}, {"receipt-schema", "--format", "execute", "../../internal/completeness/receipt.gooo"}, {"receipt-schema", "--root", "Missing", "../../internal/completeness/receipt.gooo"}} {
		var out, errOut bytes.Buffer
		if run(args, &out, &errOut) == exitOK || out.Len() != 0 {
			t.Fatal("invalid request emitted a partial schema")
		}
	}
}
