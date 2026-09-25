package languagesyntax

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDecodeRegistryRejectsDuplicateJSONKeys(t *testing.T) {
	raw, err := json.Marshal(expectedRegistry())
	if err != nil {
		t.Fatal(err)
	}
	const prefix = `{"schema":`
	if !bytes.HasPrefix(raw, []byte(prefix)) {
		t.Fatalf("registry JSON does not start with schema: %s", raw)
	}
	raw = bytes.Replace(raw, []byte(prefix), []byte(`{"schema":"sha256:duplicate-invalid","schema":`), 1)
	if _, err := decodeRegistry(raw); err == nil {
		t.Fatal("registry with duplicate JSON keys was accepted")
	}
}
