package generation

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestObservationDiagnosticsRoundTrip(t *testing.T) {
	bundle := diagnosticObservationBundle()
	payload, err := EncodeObservationBundle(bundle)
	if err != nil {
		t.Fatal(err)
	}
	var decoded OperationObservationBundle
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, bundle) {
		t.Fatal("diagnostic observation changed during canonical roundtrip")
	}
}

func TestObservationDiagnosticsBindReplayDigest(t *testing.T) {
	original := diagnosticObservationBundle()
	changed := SealObservationBundle(original)
	changed.Failures[0].Diagnostics[0] = "evidence=type-string"
	changed = SealObservationBundle(changed)
	if original.Failures[0].Diagnostics[0] != "evidence=types-check" {
		t.Fatal("sealing aliased the original diagnostic payload")
	}
	if changed.BundleDigest == original.BundleDigest || changed.ReplayDigest == original.ReplayDigest {
		t.Fatal("changed diagnostic content was not bound into both digests")
	}
	if changed.Failures[0].Decision != "UNKNOWN" || changed.Failures[0].NextOperation != "restore-type-evidence" {
		t.Fatal("diagnostics changed the causal decision")
	}
}

func diagnosticObservationBundle() OperationObservationBundle {
	return SealObservationBundle(OperationObservationBundle{
		Failures: []ObservationFailure{{
			ActionIndicatorID: "diagnostic-action",
			Decision:          "UNKNOWN", Stage: "derive-recipe", Step: "type-check-suffix",
			Reason: "TYPE_EVIDENCE_MISSING", UnknownClass: "DIRECT_MISSING",
			NextOperation: "restore-type-evidence", BlockedBy: []string{},
			Diagnostics: []string{"evidence=types-check", "unresolved-identifiers=value"},
		}},
		ObservationTotal: 1,
	})
}
