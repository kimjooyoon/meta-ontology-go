package publicpartialreuse

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestReadReceiptClassifiesContentFailure(t *testing.T) {
	valid := receiptFixtureBytes(t)
	suffix := func(value string) []byte {
		return append(append([]byte(nil), valid...), value...)
	}
	input := receiptEvaluationFixture(t)
	badSeal := input.Receipts["orders"]
	badSeal.ReceiptID = cache.HashBytes([]byte("different-test-seal")).String()
	badSealJSON, err := json.Marshal(badSeal)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		data      []byte
		missing   bool
		directory bool
		wantError bool
		invalid   bool
	}{
		{name: "valid", data: valid},
		{name: "whitespace_is_not_tampering", data: suffix("\n\t ")},
		{name: "second_json_value", data: suffix("\n{}"), wantError: true, invalid: true},
		{name: "trailing_garbage", data: suffix("!"), wantError: true, invalid: true},
		{name: "malformed_json", data: []byte("{"), wantError: true, invalid: true},
		{name: "missing_content_fields", data: []byte("{}"), wantError: true, invalid: true},
		{name: "wrong_json_shape", data: []byte("[]"), wantError: true, invalid: true},
		{name: "invalid_content_seal", data: badSealJSON, wantError: true, invalid: true},
		{name: "missing_file", missing: true, wantError: true},
		{name: "non_regular_input", directory: true, wantError: true},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "receipt.json")
			if item.directory {
				if err := os.Mkdir(filename, 0o700); err != nil {
					t.Fatal(err)
				}
			} else if !item.missing {
				if err := os.WriteFile(filename, item.data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := ReadReceipt(filename)
			if (err != nil) != item.wantError {
				t.Fatalf("ReadReceipt error=%v, want error=%v", err, item.wantError)
			}
			if errors.Is(err, ErrInvalidReceipt) != item.invalid {
				t.Fatalf("content refutation=%v, want %v: %v", errors.Is(err, ErrInvalidReceipt), item.invalid, err)
			}
		})
	}
}
