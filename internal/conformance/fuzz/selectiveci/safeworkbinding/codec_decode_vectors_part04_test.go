package safeworkbinding

import (
	"testing"
)

func TestDecodeJSON_UnknownIsolation(t *testing.T) {
	keys := []string{
		"expected",
		"expected_label",
		"want",
		"legacy_work_id",
		"result_digest",
		"replay_digest",
	}
	values := []string{
		`"ignored"`,
		"null",
		"{}",
	}
	want := decodeResultWant{
		decision:          DecisionFailClosed,
		reason:            ReasonUnknownField,
		fault:             ReasonUnknownField,
		fullSuiteRequired: false,
		resultDigest:      "sha256:1ada140e4d914ab1ab15570deb11e423a7e454b81c2bbf84512454b642ecfa02",
		resultFrameLength: 292,
	}
	for _, key := range keys {
		for _, value := range values {
			t.Run(key+"/"+value, func(t *testing.T) {
				binding, result := DecodeJSON(decodeDocumentWithUnknown(key, value))
				check(t, binding == (SafeWorkBinding{}), "unknown binding")
				checkDecodeResult(t, result, want)
			})
		}
	}
}
