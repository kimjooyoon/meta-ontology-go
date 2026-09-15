package generation

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestLinearReconciliationNormalClosesExactCandidate(t *testing.T) {
	input := reconciliationFixture(t)
	receipt := ReconcileLinearTree(input)
	if receipt.Decision != ReconciliationClosed || receipt.Reason != "EXACT_LINEAR_TREE_RECONCILIATION_CLOSED" {
		t.Fatalf("unexpected normal receipt: %+v", receipt)
	}
	if receipt.Unknown != nil || len(receipt.Refuted) != 0 || receipt.TreeDigest != input.Candidate.ExpectedTreeDigest {
		t.Fatalf("normal receipt retained unexpected failure evidence: %+v", receipt)
	}
	if receipt.Metrics != (ReconciliationMetrics{OperationRequests: 1, OperationResults: 1, ReplayComparisons: 1}) {
		t.Fatalf("unexpected normal metrics: %+v", receipt.Metrics)
	}
}

func TestLinearReconciliationMissingAuthorizationIsUnknownWithExactFields(t *testing.T) {
	input := reconciliationFixture(t)
	input.Authorization = nil
	receipt := ReconcileLinearTree(input)
	if receipt.Decision != ReconciliationUnknown || receipt.Unknown == nil || receipt.Unknown.Reason != "OWNER_AUTHORIZATION_MISSING" {
		t.Fatalf("unexpected unknown receipt: %+v", receipt)
	}
	encoded, err := json.Marshal(receipt.Unknown)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	expected := append([]string(nil), ReconciliationUnknownFields[:]...)
	sort.Strings(expected)
	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("unknown evidence fields changed: got %v want %v", keys, expected)
	}
}

func TestLinearReconciliationRejectsNonLinearOrNonEquivalentTree(t *testing.T) {
	input := reconciliationFixture(t)
	input.Topology.Status = "diverged"
	input.Trees.Reconciled = &ReconciliationTree{{Path: "main.go", Mode: "100644", Type: "blob", SHA: strings.Repeat("b", 40)}}
	receipt := ReconcileLinearTree(input)
	if receipt.Decision != ReconciliationRefuted || len(receipt.Refuted) < 2 {
		t.Fatalf("expected topology and tree refutations: %+v", receipt)
	}
	if receipt.Reason != "NON_LINEAR_TOPOLOGY" {
		t.Fatalf("refuted precedence did not retain first exact reason: %+v", receipt)
	}
}

func TestLinearReconciliationReplayMatchClosesAndCollisionRefutes(t *testing.T) {
	input := reconciliationFixture(t)
	input.Replay.PreviousIdentity = input.Replay.Identity
	input.Replay.PreviousRequestDigest = input.Replay.CurrentRequestDigest
	matched := ReconcileLinearTree(input)
	if matched.Decision != ReconciliationClosed || matched.Metrics.ReplayMismatches != 0 {
		t.Fatalf("same replay did not close: %+v", matched)
	}

	input.Replay.PreviousRequestDigest = strings.Repeat("c", 64)
	collision := ReconcileLinearTree(input)
	if collision.Decision != ReconciliationRefuted || collision.Reason != "REPLAY_COLLISION" || collision.Metrics.ReplayMismatches != 1 {
		t.Fatalf("replay collision was not refuted exactly: %+v", collision)
	}
}

func TestLinearReconciliationAuthorizationMustBeFreshAndMainTargeted(t *testing.T) {
	input := reconciliationFixture(t)
	input.Authorization.ExpiresAt = "2026-09-02T00:00:00Z"
	receipt := ReconcileLinearTree(input)
	if receipt.Decision != ReconciliationRefuted || receipt.Reason != "AUTHORIZATION_NOT_FRESH" {
		t.Fatalf("stale authorization was not refuted: %+v", receipt)
	}

	input = reconciliationFixture(t)
	input.Authorization.TargetBranch = "dev"
	receipt = ReconcileLinearTree(input)
	if receipt.Decision != ReconciliationRefuted || receipt.Reason != "OWNER_AUTHORIZATION_MISMATCH" {
		t.Fatalf("wrong target authorization was not refuted: %+v", receipt)
	}
}

func reconciliationFixture(t *testing.T) ReconciliationInput {
	t.Helper()
	tree := ReconciliationTree{{Path: "main.go", Mode: "100644", Type: "blob", SHA: strings.Repeat("a", 40)}}
	treeDigest, err := ReconciliationTreeDigest(tree)
	if err != nil {
		t.Fatal(err)
	}
	candidate := ReconciliationCandidate{
		PullRequest:        651,
		Repository:         ReconciliationRepository,
		BaseBranch:         "main",
		BaseSHA:            strings.Repeat("1", 40),
		HeadBranch:         "dev",
		HeadSHA:            strings.Repeat("2", 40),
		MergeBaseSHA:       strings.Repeat("1", 40),
		ReplayIdentity:     "main-target:651",
		ExpectedTreeDigest: treeDigest,
	}
	input := ReconciliationInput{
		Candidate: candidate,
		Topology: ReconciliationTopology{
			MainBeforeSHA: candidate.BaseSHA, MainAfterSHA: candidate.BaseSHA,
			DevBeforeSHA: candidate.HeadSHA, DevAfterSHA: candidate.HeadSHA,
			MergeBaseSHA: candidate.MergeBaseSHA, Status: "ahead", AheadBy: 1, BehindBy: 0,
		},
		Trees:     ReconciliationTreeEvidence{SourceBefore: &tree, SourceAfter: &tree, Reconciled: &tree},
		Mutations: ReconciliationMutationEvidence{},
		Now:       "2026-09-03T01:00:00Z",
	}
	owner := ReconciliationOwnerIdentity()
	input.Authorization = &ReconciliationAuthorization{
		State: "AUTHORIZED", TargetBranch: "main", OwnerSelection: owner, Actor: owner,
		Candidate: candidate, CandidateDigest: ReconciliationCandidateDigest(candidate),
		Nonce: "linear-reconciliation-nonce", IssuedAt: "2026-09-03T00:00:00Z", ExpiresAt: "2026-09-04T00:00:00Z",
		OneUse: true, Reusable: false, UseCount: 0, ReuseAttempts: 0,
		ProtectionMutation: false, ForcePush: false, RepositoryWritesBefore: 0,
	}
	input.Replay = &ReconciliationReplay{
		Identity: candidate.ReplayIdentity, CurrentRequestDigest: ReconciliationCandidateDigest(candidate),
	}
	return input
}
