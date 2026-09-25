package selectiveci

import (
	"testing"
)

func TestBaselineMultiRootCoverageIsLaunderedByCoveredRoot(t *testing.T) {
	input := completeInput()
	input.Base.Files = append(input.Base.Files, SnapshotFile{
		Path:        "billing/customer.gooo",
		BlobDigest:  digest("customer-base"),
		SemanticIDs: []string{"urn:selectiveci:entity/customer"},
	})
	input.Head.Files = append(input.Head.Files, SnapshotFile{
		Path:        "billing/customer.gooo",
		BlobDigest:  digest("customer-head"),
		SemanticIDs: []string{"urn:selectiveci:entity/customer"},
	})
	input.Base.Digest = input.Base.ComputedDigest()
	input.Head.Digest = input.Head.ComputedDigest()
	for index := range input.Receipts {
		input.Receipts[index].SnapshotDigest = input.Head.Digest
	}

	got := Plan(input)
	if got.Status != StatusFullSuiteFallback || got.ReasonCode != "MISSING_OBLIGATION" {
		t.Fatalf("multi-root status = %s/%s, want FULL_SUITE_FALLBACK/MISSING_OBLIGATION", got.Status, got.ReasonCode)
	}
	if len(got.SelectedCommandIDs)+len(got.SelectedGuardCommandIDs)+len(got.SelectedWorkIDs)+len(got.ResourceReceiptDigests)+len(got.ProvenancePathIDs) != 0 {
		t.Fatalf("uncovered root retained selection evidence: %#v", got)
	}
}
func TestBaselineZeroObligationSingleRootFallsBack(t *testing.T) {
	input := completeInput()
	input.Registry.Obligations = []ObligationBinding{}
	input.Registry.Nodes = nodesWithoutObligation(input.Registry.Nodes)
	input.Registry.Digest = input.Registry.ComputedDigest()

	got := Plan(input)
	if got.Status != StatusFullSuiteFallback || got.ReasonCode != "MISSING_OBLIGATION" {
		t.Fatalf("zero-obligation status = %s/%s, want FULL_SUITE_FALLBACK/MISSING_OBLIGATION", got.Status, got.ReasonCode)
	}
}
func TestPlanCoverageCommandFailuresClearSelectionEvidence(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		reason string
	}{
		{"missing command", func(input *Input) {
			input.Registry.Obligations[0].CommandIDs = []string{}
			input.Registry.Digest = input.Registry.ComputedDigest()
		}, ReasonMissingCommand},
		{"dangling command", func(input *Input) {
			input.Registry.Obligations[0].CommandIDs = []string{"urn:selectiveci:command/missing"}
			input.Registry.Digest = input.Registry.ComputedDigest()
		}, ReasonDanglingCommand},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := completeInput()
			test.mutate(&input)
			got := Plan(input)
			if got.Status != StatusFullSuiteFallback || got.ReasonCode != test.reason {
				t.Fatalf("plan = %s/%s, want FULL_SUITE_FALLBACK/%s", got.Status, got.ReasonCode, test.reason)
			}
			if len(got.SelectedCommandIDs)+len(got.SelectedGuardCommandIDs)+len(got.SelectedWorkIDs)+len(got.ResourceReceiptDigests)+len(got.ProvenancePathIDs) != 0 {
				t.Fatalf("plan retained selection evidence: %#v", got)
			}
		})
	}
}
