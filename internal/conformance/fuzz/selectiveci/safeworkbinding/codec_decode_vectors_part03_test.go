package safeworkbinding

import (
	"testing"
)

func checkDecodeResult(t *testing.T, result ParseResult, want decodeResultWant) {
	t.Helper()
	check(t, result.Decision == want.decision, "decision")
	check(t, result.Reason == want.reason, "reason")
	if want.emptyFaults {
		check(t, result.Faults != nil && len(result.Faults) == 0, "empty faults")
	} else {
		check(t, result.Faults != nil && len(result.Faults) == 1, "singleton fault")
		check(t, result.Faults[0] == want.fault, "fault reason")
	}
	check(t, result.FullSuiteRequired == want.fullSuiteRequired, "full suite")
	check(t, !result.ExecutionAuthorized, "execution authorization")
	check(t, result.EnforcementEffect == EnforcementEffectNoEffect, "enforcement effect")
	check(t, result.ResultDigest == want.resultDigest, "result digest")
	check(t, result.ReplayDigest == want.replayDigest, "replay digest")
}
func decodeDocumentWithUnknown(key, value string) []byte {
	document := append([]byte(nil), decodeDocument(nil, nil)...)
	document = document[:len(document)-1]
	document = append(document, ',', '"')
	document = append(document, key...)
	document = append(document, '"', ':')
	document = append(document, value...)
	document = append(document, '}')
	return document
}
func TestDecodeJSON_ResultReplayVectors(t *testing.T) {
	for _, vector := range decodeVectors {
		t.Run(vector.name, func(t *testing.T) {
			binding, result := DecodeJSON(vector.input)
			check(t, binding == vector.binding, "binding")
			checkDecodeResult(t, result, vector.want)
			frame, ok := resultFrame(result)
			check(t, ok, "result frame")
			check(t, len(frame) == vector.want.resultFrameLength, "result frame length")
		})
	}
	binding, result := DecodeJSON(decodeVectors[0].input)
	check(t, len(replayFrame(binding.BindingDigest, result.ResultDigest)) == 252, "replay frame length")
}
