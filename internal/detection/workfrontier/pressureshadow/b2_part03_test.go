package pressureshadow

import (
	"strings"
	"testing"
)

func TestB2UpstreamPropagationAndStrictWire(t *testing.T) {
	for _, test := range []struct {
		name     string
		mutate   func(*Input)
		decision Decision
		reason   Reason
	}{
		{name: "unknown", mutate: func(input *Input) {
			input.Selector.SnapshotDigest = ""
		}, decision: DecisionUnknown, reason: ReasonUpstreamUnknown},
		{name: "fail closed", mutate: func(input *Input) {
			input.Selector.Paths[0].StableID = "path a"
		}, decision: DecisionFailClosed, reason: ReasonUpstreamFailClosed},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := b2Input(t)
			test.mutate(&input)
			upstream := ValidateB1(input)
			got := ValidateB2(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				got.InputDigest != upstream.InputDigest || got.UpstreamResultDigest != upstream.ResultDigest ||
				len(got.MissingRequiredPressureRecordIDs) != 0 || len(got.MissingSelectorPressureIDs) != 0 ||
				len(got.UnregisteredPressureRecordIDs) != 0 {
				t.Fatalf("upstream propagation = %#v", got)
			}
		})
	}
	for _, raw := range []string{
		strings.Replace(b2RawInput, `"schema":`, `"expected_label":"PASS", "schema":`, 1),
		strings.Replace(b2RawInput, `"schema":`, `"schema":"duplicate", "schema":`, 1),
		b2RawInput + `{}`,
	} {
		got := ValidateB2Bytes([]byte(raw))
		if got.Decision != DecisionFailClosed || got.Reason != ReasonUpstreamFailClosed ||
			len(got.MissingRequiredPressureRecordIDs) != 0 {
			t.Fatalf("strict wire result = %#v", got)
		}
	}
}
