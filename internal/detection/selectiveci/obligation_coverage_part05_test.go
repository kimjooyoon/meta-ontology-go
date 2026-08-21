package selectiveci

import (
	"math"
	"testing"
)

func TestObligationCoverageNoChangePermutationAndCanonicalOutput(t *testing.T) {
	input := coverageInputFromPlanInput(t, completeInput(), []string{}...)
	got := ObserveObligationCoverage(input)
	if got.Decision != CoverageDecisionExact || got.Reason != CoverageReasonNoChange || got.DeterministicWorkUnits != 0 {
		t.Fatalf("no-change coverage = %#v", got)
	}
	permuted := input
	permuted.ChangedRootIDs = []string{}
	permuted.Graph.Nodes = reverseImpactNodes(permuted.Graph.Nodes)
	permuted.Graph.Edges = reverseImpactEdges(permuted.Graph.Edges)
	permuted.Registry.Obligations = reverseObligationsForCoverage(permuted.Registry.Obligations)
	permuted.Registry.Commands = reverseCommandsForCoverage(permuted.Registry.Commands)
	permuted.Registry.Digest = permuted.Registry.ComputedDigest()
	permuted.Graph.RegistryDigest = permuted.Registry.Digest
	left, err := EncodeObligationCoverageJSON(got)
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeObligationCoverageJSON(ObserveObligationCoverage(permuted))
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) {
		t.Fatalf("permutation changed coverage output:\n%s\n%s", left, right)
	}
	inputJSON, err := EncodeObligationCoverageInputJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeObligationCoverageInputJSON(inputJSON); err != nil {
		t.Fatalf("input JSON round trip failed: %v", err)
	}
	if _, err := DecodeObligationCoverageInputJSON([]byte(`{"schema_version":"gooo/selective-ci-obligation-coverage/v1","schema_version":"gooo/selective-ci-obligation-coverage/v1"}`)); err == nil {
		t.Fatal("duplicate input fields were accepted")
	}
}
func TestObligationCoverageWorkOverflowAndReasonLabels(t *testing.T) {
	if _, ok := coverageWorkUnits(math.MaxUint64, 1, 0); ok {
		t.Fatal("root work overflow was accepted")
	}
	if _, ok := coverageWorkUnits(1, math.MaxUint64, 1); ok {
		t.Fatal("obligation work overflow was accepted")
	}
	if _, ok := coverageWorkUnits(1, 1, math.MaxUint64); ok {
		t.Fatal("command-binding work overflow was accepted")
	}
	if _, err := DecodeObligationCoverageJSON([]byte(`{"schema_version":"gooo/selective-ci-obligation-coverage/v1","decision":"UNKNOWN","reason":"not-a-reason","full_suite_required":true,"changed_root_count":0,"covered_changed_root_count":0,"uncovered_changed_root_count":0,"required_obligation_count":0,"bound_command_count":0,"deterministic_work_units":0,"uncovered_root_ids":[],"required_obligation_ids":[],"graph_digest":"","registry_digest":"","snapshot_digest":"","input_digest":"","output_digest":""}`)); err == nil {
		t.Fatal("unknown reason label was accepted")
	}
}
func assertMissingCoverage(t *testing.T, got ObligationCoverageResult, reason CoverageReason, roots, covered, uncovered uint64) {
	t.Helper()
	if got.Decision != CoverageDecisionUnknown || got.Reason != reason || !got.FullSuiteRequired {
		t.Fatalf("coverage = %#v, want UNKNOWN/%s", got, reason)
	}
	if got.ChangedRootCount != roots || got.CoveredChangedRootCount != covered || got.UncoveredChangedRootCount != uncovered {
		t.Fatalf("root counts = %#v", got)
	}
	if len(got.RequiredObligationIDs) != 0 {
		t.Fatalf("unknown coverage exposed required IDs: %v", got.RequiredObligationIDs)
	}
}
