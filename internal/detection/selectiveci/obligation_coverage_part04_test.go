package selectiveci

import (
	"testing"
)

func TestObligationCoverageCommandAndInputReasons(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ObligationCoverageInput)
		reason CoverageReason
	}{
		{"zero command", func(input *ObligationCoverageInput) {
			input.Registry.Obligations[0].CommandIDs = []string{}
			input.Registry.Digest = input.Registry.ComputedDigest()
			input.Graph.RegistryDigest = input.Registry.Digest
		}, CoverageReasonMissingCommand},
		{"dangling command", func(input *ObligationCoverageInput) {
			input.Registry.Obligations[0].CommandIDs = []string{"urn:selectiveci:command/missing"}
			input.Registry.Digest = input.Registry.ComputedDigest()
			input.Graph.RegistryDigest = input.Registry.Digest
		}, CoverageReasonDanglingCommand},
		{"unknown root", func(input *ObligationCoverageInput) {
			input.ChangedRootIDs = []string{"urn:selectiveci:entity/missing"}
		}, CoverageReasonUnknownRoot},
		{"unsupported schema", func(input *ObligationCoverageInput) {
			input.SchemaVersion = "gooo/selective-ci-obligation-coverage/v0"
		}, CoverageReasonUnsupportedSchema},
		{"duplicate root", func(input *ObligationCoverageInput) {
			input.ChangedRootIDs = append(input.ChangedRootIDs, input.ChangedRootIDs[0])
		}, CoverageReasonDuplicateRoot},
		{"stale graph", func(input *ObligationCoverageInput) { input.Graph.RegistryDigest = digest("stale-graph") }, CoverageReasonStaleGraph},
		{"stale registry", func(input *ObligationCoverageInput) {
			input.Registry.Digest = digest("stale-registry")
		}, CoverageReasonStaleRegistry},
		{"stale snapshot", func(input *ObligationCoverageInput) { input.SnapshotDigest = digest("stale-snapshot") }, CoverageReasonStaleSnapshot},
		{"invalid graph", func(input *ObligationCoverageInput) { input.Graph.Version = "gooo/impact-graph/v0" }, CoverageReasonInvalidGraph},
		{"missing snapshot", func(input *ObligationCoverageInput) { input.SnapshotDigest = "" }, CoverageReasonInvalidSnapshot},
		{"missing roots", func(input *ObligationCoverageInput) { input.ChangedRootIDs = nil }, CoverageReasonMissingInput},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := coverageInputFromPlanInput(t, completeInput(), "urn:selectiveci:entity/order")
			test.mutate(&input)
			got := ObserveObligationCoverage(input)
			if got.Decision != CoverageDecisionUnknown || got.Reason != test.reason || len(got.RequiredObligationIDs) != 0 || !got.FullSuiteRequired {
				t.Fatalf("coverage = %#v, want UNKNOWN/%s without required IDs", got, test.reason)
			}
		})
	}
}
