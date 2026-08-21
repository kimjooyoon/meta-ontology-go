package pressureshadow

import (
	"reflect"
	"strings"
	"testing"
)

func runPathVectors(t *testing.T, cases []pathVector) {
	t.Helper()
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := a2aInput(t)
			test.mutate(&input)
			assertPathVector(t, Validate(input), test.decision, test.reason,
				test.missing, test.orphan, test.missingBinding, test.mismatch)
		})
	}
}
func TestValidateBytesRejectsInvalidWire(t *testing.T) {
	cases := []struct {
		name, raw, inputDigest, resultDigest, replayDigest string
	}{
		{
			name: "unknown key",
			raw: strings.Replace(a2aRawInput, `"schema":`,
				`"expected_reason":"PASS", "schema":`, 1),
			inputDigest:  "sha256:80daab8a5314b24fb407b75927933558f264f4a23ea51c8ec8890da4efa0a67f",
			resultDigest: "sha256:c65d1f50658330fb78b346e3d1ced7ca1c54e107b6325cc139337532b2250984",
			replayDigest: "sha256:9c75f7d8607039b95c5d9a9368537f7f4926a9be84ec81d834b2c4fe13ffb816",
		},
		{
			name: "duplicate key",
			raw: strings.Replace(a2aRawInput, `"schema":`,
				`"schema":"duplicate", "schema":`, 1),
			inputDigest:  "sha256:4b223fa4b6fafeec641faabaaa1b1c4aa044c2988c84223f20528bda2fa4b82a",
			resultDigest: "sha256:6de7a3f27c582df3027c8da8c7d9781b867ffe4c3a0c8fc7e9a6ce8cd4e7a202",
			replayDigest: "sha256:feb305dfe88becf7dce326b3bd68d5951fda147520ce827e13e1f8f312b0c77b",
		},
		{
			name: "trailing value", raw: a2aRawInput + `{}`,
			inputDigest:  "sha256:e68ff81fcb90db91e96019370d3299ef9f6a37f9cae22ab285ca60b098515c3c",
			resultDigest: "sha256:668a6a0d1c74e6095b4248cfb20c66ef61e6c5ac780c3974fa973aabf08134c9",
			replayDigest: "sha256:f3b3c7d91f25658bb0ec8e1763f5b6a79eafa3ab031e73467dffac845e56524b",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			want := invalidWireResult(test.inputDigest, test.resultDigest, test.replayDigest)
			if got := ValidateBytes([]byte(test.raw)); !reflect.DeepEqual(got, want) {
				t.Fatalf("invalid wire result = %#v, want %#v", got, want)
			}
		})
	}
}
func invalidWireResult(inputDigest, resultDigest, replayDigest string) Result {
	return Result{
		Schema: SchemaVersion, InputDigest: inputDigest,
		Decision: DecisionFailClosed, Reason: ReasonInvalidInput,
		MissingPathIDs: []string{}, OrphanPathIDs: []string{},
		MissingBindingPathIDs: []string{}, BindingMismatchPathIDs: []string{},
		EnforcementEffect: EnforcementNoEffect,
		ResultDigest:      resultDigest, ReplayDigest: replayDigest,
	}
}
func TestResultDigestBindsExpectedLabels(t *testing.T) {
	result := Validate(a2aInput(t))
	mutated := result
	mutated.Decision, mutated.Reason = DecisionUnknown, ReasonBindingMismatch
	if mutated.InputDigest != result.InputDigest || CanonicalResultDigest(mutated) == result.ResultDigest {
		t.Fatal("expected-label mutation was not bound to result digest")
	}
}
