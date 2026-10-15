package bodyexecution

import (
	"strings"
	"testing"
)

func TestTextInputChecksOriginalEscapesBeforeTypedKeys(t *testing.T) {
	// A replacement decoder could collapse the malformed key onto the valid key.
	raw := []byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"\ud80d":1,"�":2}]}`)
	suite, err := DecodeCompositionInputs(raw)
	if err == nil || !strings.Contains(err.Error(), "UNPAIRED_HIGH") || len(suite.Cases) != 0 {
		t.Fatal("typed key decoding preceded original escape policy", err)
	}
	var output struct {
		Label string `json:"label"`
	}
	if err := decodeCompositionTextData([]byte(`{"label":"suffix-\ud80e"}`), &output, 1024); err == nil || output.Label != "" {
		t.Fatal("rejected input modified typed output", output, err)
	}
}

func TestTextInputEarlyPolicyKeepsReplacementScalarAndByteLimit(t *testing.T) {
	text, err := decodeTextScalar([]byte(`"control-\ufffd"`))
	if err != nil || text != "control-�" {
		t.Fatal("explicit replacement scalar refused", text, err)
	}
	// Escape classification precedes the decoded-size check on a malformed input.
	bad := []byte(`"` + strings.Repeat("z", 1025) + `\ud80f"`)
	if _, err := decodeTextScalar(bad); err == nil || !strings.Contains(err.Error(), "UNPAIRED_HIGH") {
		t.Fatal("original escape reason lost", err)
	}
	if _, err := decodeTextScalar([]byte(`"` + strings.Repeat("q", 1025) + `"`)); err == nil {
		t.Fatal("early escape policy bypassed decoded byte limit")
	}
}
