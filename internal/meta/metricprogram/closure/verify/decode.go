package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func decodeInput(in Input, raw []byte) (receipt, programDocument, verificationDocument, error) {
	var closure receipt
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&closure); err != nil {
		return closure, programDocument{}, verificationDocument{}, fmt.Errorf("decode closure: %w", err)
	}
	var program programDocument
	if err := json.Unmarshal(in.ProgramJSON, &program); err != nil {
		return closure, program, verificationDocument{}, fmt.Errorf("decode program: %w", err)
	}
	var verification verificationDocument
	if err := json.Unmarshal(in.VerificationJSON, &verification); err != nil {
		return closure, program, verification, fmt.Errorf("decode verification: %w", err)
	}
	return closure, program, verification, nil
}
