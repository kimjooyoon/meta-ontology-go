package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// DecodeRuntimeReceipt extracts exact shared receipt bytes from a shape-checked
// runtime envelope. It does not attest execution, permissions, or its producer.
// Consumers requiring execution proof must replay the source with Execute.
func DecodeRuntimeReceipt(data []byte) ([]byte, error) {
	var result Result
	if err := decode(data, &result, 2<<20); err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	if len(object) != 3 || object["observation"] == nil || object["parent_receipt_bytes"] == nil ||
		result.Observation.Schema != "gooo/typed-path-runtime-observation/v1" {
		return nil, fmt.Errorf("incomplete or unsupported runtime envelope")
	}
	raw := object["completeness_receipt"]
	r, err := completeness.Decode(raw)
	if err != nil {
		return nil, err
	}
	if _, err := completeness.Decode(result.ParentReceipt); err != nil {
		return nil, err
	}
	parent := digest(result.ParentReceipt)
	if r.Scope["parent_receipt_sha256"] != parent ||
		result.Observation.ParentReceiptSHA256 != parent {
		return nil, fmt.Errorf("runtime parent receipt binding differs")
	}
	if err := runtimeProfileBinding(result, r); err != nil {
		return nil, err
	}
	return bytes.Clone(raw), nil
}
