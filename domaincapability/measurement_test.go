package domaincapability

import "testing"

func TestMeasureFromBindingsKeepsUnresolvedCapabilities(t *testing.T) {
	measurement := MeasureFromBindings(
		[]string{"gooo.declaration.diff", "gooo.provenance.compare"},
		[]CapabilityBinding{{
			CapabilityID: "gooo.declaration.diff", DeclarationID: "decl-diff",
			SourceDigest: "sha256:source", EvidenceDigest: "sha256:evidence",
		}},
		3,
	)
	if measurement.ExpectedCapabilityCount != 2 || measurement.ObservedCapabilityCount != 1 || measurement.UnresolvedCapabilityCount != 1 {
		t.Fatalf("measurement = %+v", measurement)
	}
	if !measurement.EvidenceBound || measurement.ObservationCount != 3 {
		t.Fatalf("measurement lost evidence boundary: %+v", measurement)
	}
}