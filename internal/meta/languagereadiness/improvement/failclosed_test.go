package improvement

import "testing"

func TestEvaluateFailsClosedWithoutInference(t *testing.T) {
	cases := []struct {
		name   string
		reason string
		mutate func(*Snapshot)
	}{
		{name: "unknown evidence", reason: "AFTER_EVIDENCE_STATUS_UNKNOWN", mutate: func(snapshot *Snapshot) {
			snapshot.Evidence[23].Status = "LIKELY_SATISFIED"
		}},
		{name: "unresolved evidence", reason: "UNRESOLVED_EVIDENCE", mutate: func(snapshot *Snapshot) {
			snapshot.Evidence[23].Status = Unresolved
		}},
		{name: "moving registry", reason: "REGISTRY_DIGEST_MISMATCH", mutate: func(snapshot *Snapshot) {
			snapshot.RegistryDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
		}},
		{name: "moving denominator", reason: "AFTER_DENOMINATOR_INVALID", mutate: func(snapshot *Snapshot) {
			snapshot.Total++
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			after := testSnapshot(8)
			test.mutate(&after)
			got := Evaluate(testSnapshot(7), after)
			if got.Decision != LowerResolution || got.ReasonCode != test.reason {
				t.Fatalf("decision = %q, reason = %q", got.Decision, got.ReasonCode)
			}
		})
	}
}

func TestEvaluateSeparatesNoChangeAndRegression(t *testing.T) {
	noChange := Evaluate(testSnapshot(7), testSnapshot(7))
	if noChange.Decision != NoChange || noChange.CompletedDelta != 0 || noChange.BasisPointsDelta != 0 {
		t.Fatalf("no change = %#v", noChange)
	}
	regressed := Evaluate(testSnapshot(7), testSnapshot(6))
	if regressed.Decision != Regressed || regressed.CompletedDelta != -1 || regressed.Regressions != 1 {
		t.Fatalf("regression = %#v", regressed)
	}
}
