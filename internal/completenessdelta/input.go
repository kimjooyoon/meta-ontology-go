package completenessdelta

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

const MaxInputBytes = 2 << 20

// receiptBytes selects exactly one supported producer path. Envelopes are
// shape-checked observations, not attestations of execution or model calls.
func receiptBytes(data []byte) ([]byte, error) {
	if err := completeness.CheckJSON(data, MaxInputBytes); err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return nil, fmt.Errorf("input must be a receipt or supported producer object")
	}
	if _, ok := object["schema"]; ok {
		if _, err := completeness.Decode(data); err != nil {
			return nil, err
		}
		return bytes.Clone(data), nil
	}
	if _, ok := object["report"]; ok {
		_, receipt, err := bodyexecution.DecodeGeneration(data)
		return receipt, err
	}
	if _, ok := object["completeness_receipt"]; ok {
		return bodyexecution.DecodeRuntimeReceipt(data)
	}
	return nil, fmt.Errorf("no supported receipt path")
}

func record(data []byte) map[string]any {
	// Called only after the complete receipt was decoded and validated.
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var r map[string]any
	_ = d.Decode(&r)
	return r
}
