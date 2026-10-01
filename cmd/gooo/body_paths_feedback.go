package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func decodePathCIHint(raw []byte) (*pathplan.CIHint, error) {
	if len(raw) == 0 || len(raw) > 512 {
		return nil, fmt.Errorf("typed path CI hint must be at most 512 bytes")
	}
	if err := decision.RejectDuplicateJSONKeys(raw); err != nil {
		return nil, fmt.Errorf("typed path CI hint JSON is invalid")
	}
	var hint pathplan.CIHint
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&hint); err != nil {
		return nil, fmt.Errorf("typed path CI hint fields are invalid")
	}
	if err := hint.Validate(); err != nil {
		return nil, err
	}
	return &hint, nil
}
