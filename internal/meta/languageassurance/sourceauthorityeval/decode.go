package sourceauthorityeval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func Decode(raw []byte) (Bundle, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var bundle Bundle
	if err := decoder.Decode(&bundle); err != nil {
		return Bundle{}, fmt.Errorf("decode source authority evidence: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Bundle{}, fmt.Errorf("decode source authority evidence: trailing value")
	}
	return bundle, nil
}

func Observe(raw []byte) Report {
	bundle, err := Decode(raw)
	if err == nil {
		return Evaluate(bundle)
	}
	report := newReport(Bundle{})
	setOutcome(&report, "UNKNOWN", "INVARIANT_ONLY", "BLOCK",
		"SOURCE_AUTHORITY_EVIDENCE_UNKNOWN")
	return seal(report)
}
