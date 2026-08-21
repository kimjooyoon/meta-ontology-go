package coupling

import (
	"encoding/json"
	"testing"
)

func TestDecodeResultRejectsForgedClosedAlgebra(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	valid := Evaluate(fixture.input, fixture.authorityContext)
	cases := map[string]func(*Result){
		"pass with reason": func(result *Result) { result.Reasons = []Reason{{Code: ReasonDigestMismatch, Detail: "x"}} },
		"pass full suite":  func(result *Result) { result.FullSuiteRequired = true },
		"unknown accepted": func(result *Result) {
			result.Status = StatusUnknown
			result.Reasons = []Reason{{Code: ReasonDigestMismatch, Detail: "x"}}
		},
		"fail no reason": func(result *Result) {
			result.Status = StatusFailClosed
			result.Reasons = nil
			result.AcceptedSurfaceIDs = nil
		},
		"duplicate ID": func(result *Result) {
			result.AcceptedSurfaceIDs = append(result.AcceptedSurfaceIDs, result.AcceptedSurfaceIDs[0])
		},
		"duplicate reason": func(result *Result) {
			result.Status = StatusUnknown
			result.AcceptedSurfaceIDs = nil
			result.FullSuiteRequired = true
			result.Reasons = []Reason{{Code: ReasonDigestMismatch, Detail: "x"}, {Code: ReasonDigestMismatch, Detail: "x"}}
		},
		"unknown reason": func(result *Result) {
			result.Status = StatusUnknown
			result.AcceptedSurfaceIDs = nil
			result.FullSuiteRequired = true
			result.Reasons = []Reason{{Code: "N/A", Detail: "x"}}
		},
		"unknown dimension": func(result *Result) { result.Observation.CPU = CountDimension{Value: 1} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			forged := valid
			mutate(&forged)
			forged.Digest = stableDigest(resultCanonical(forged))
			data, err := json.Marshal(forged)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeResult(data); err == nil {
				t.Fatal("accepted forged result")
			}
		})
	}
}
