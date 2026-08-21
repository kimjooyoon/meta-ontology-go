package couplingexplain

import (
	"encoding/json"
	"testing"
)

func digest(value string) string { return DigestBytes([]byte(value)) }
func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func TestNoLinkReasonsAreClosed(t *testing.T) {
	for _, value := range []LinkReason{ReasonAmbiguous, ReasonStale, ReasonUnregistered, ReasonMissing, ReasonNotVerified} {
		if !validReason(value) {
			t.Fatal(value)
		}
	}
}
