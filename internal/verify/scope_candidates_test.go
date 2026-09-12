package verify

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestUnknownAgentBranchHintsPreserveDenial(t *testing.T) {
	branch := "agent/unregistered-scope-candidate-witness"
	paths := []string{
		"internal/bidir/reconcile_rollback_attempt_test.go",
		"internal/bidir/facts_part05.go",
		"internal/bidir/bx_evidence_part03.go",
		"internal/bidir/reconcile_part01.go",
	}
	before := append([]string(nil), paths...)
	err := CheckPathScopeForBranch(paths, branch)
	var failure *UnknownAgentBranchError
	if !errors.As(err, &failure) || failure.RequestedBranch() != branch {
		t.Fatalf("unknown owner did not retain its typed denial: %v", err)
	}
	candidates := failure.CandidateBranches()
	if !slices.Contains(candidates, "agent/bidir") || !slices.Contains(candidates, "agent/bidir-followup") {
		t.Fatalf("declared bidir owners are missing: %v", candidates)
	}
	if slices.Contains(candidates, branch) || slices.Contains(candidates, "agent/ci-workflow") {
		t.Fatalf("out-of-scope or unknown owner became a candidate: %v", candidates)
	}
	for _, candidate := range candidates {
		if err := CheckPathScopeForBranch(paths, candidate); err != nil {
			t.Fatalf("candidate does not own every path: %s: %v", candidate, err)
		}
	}
	if !strings.Contains(failure.Error(), "no paths are allowed") || !strings.Contains(failure.Error(), "scope-only") {
		t.Fatalf("diagnostic lost denial or its limited scope: %v", failure)
	}
	if !reflect.DeepEqual(paths, before) {
		t.Fatal("candidate discovery mutated caller paths")
	}
	candidates[0] = "forged-owner"
	if slices.Contains(failure.CandidateBranches(), "forged-owner") {
		t.Fatal("candidate accessor exposed mutable error state")
	}
}

func TestScopeCandidatesRequireAllPathsAndReplayInOrder(t *testing.T) {
	bidir, cli := "internal/bidir/facts_part05.go", "cmd/gooo/main.go"
	paths := []string{bidir, cli, bidir}
	got := CandidateScopeBranches(paths)
	replay := CandidateScopeBranches([]string{cli, bidir})
	if !reflect.DeepEqual(got, replay) || !slices.IsSorted(got) {
		t.Fatalf("scope candidates depend on input order or duplicates: %v / %v", got, replay)
	}
	if slices.Contains(got, "agent/bidir") || slices.Contains(got, "agent/cli") {
		t.Fatalf("a partial owner was suggested for the whole change: %v", got)
	}
	for _, branch := range got {
		if err := CheckPathScopeForBranch(paths, branch); err != nil {
			t.Fatalf("candidate bypassed the existing checker: %s: %v", branch, err)
		}
	}
	if !slices.Contains(CandidateScopeBranches([]string{cli}), "agent/cli") {
		t.Fatal("candidate discovery was hard-coded to the bidir incident")
	}
}

func TestScopeCandidatesDoNotGuessFromMissingOrMalformedPaths(t *testing.T) {
	cases := [][]string{
		nil,
		{},
		{"."},
		{".", "internal/bidir/facts_part05.go"},
		{""},
		{"../internal/bidir/facts_part05.go"},
		{"internal/bidir/../bidir/facts_part05.go"},
		{"/internal/bidir/facts_part05.go"},
		{"internal\\bidir\\facts_part05.go"},
	}
	for _, paths := range cases {
		candidates := CandidateScopeBranches(paths)
		if candidates == nil || len(candidates) != 0 {
			t.Fatalf("incomplete or malformed scope produced candidates: %q: %v", paths, candidates)
		}
	}
}

func TestScopeCandidatesKeepDirectoryComponentBoundary(t *testing.T) {
	candidates := CandidateScopeBranches([]string{"internal/bidir-other/facts.go"})
	if slices.Contains(candidates, "agent/bidir") || slices.Contains(candidates, "agent/bidir-followup") {
		t.Fatalf("a directory-name prefix escaped its component boundary: %v", candidates)
	}
}

func TestScopeCandidateHintsLeaveKnownBranchAdmissionUnchanged(t *testing.T) {
	if err := CheckPathScopeForBranch([]string{"internal/bidir/facts_part05.go"}, "agent/bidir"); err != nil {
		t.Fatalf("existing owner was rejected: %v", err)
	}
	err := CheckPathScopeForBranch([]string{"cmd/gooo/main.go"}, "agent/bidir")
	if err == nil {
		t.Fatal("known owner acquired an out-of-scope path")
	}
	var unknown *UnknownAgentBranchError
	if errors.As(err, &unknown) {
		t.Fatalf("a known owner's path denial was relabeled unknown: %v", err)
	}
}

func TestUnknownBranchWithoutPathsKeepsExistingDiagnostic(t *testing.T) {
	err := CheckPathScopeForBranch(nil, "agent/unregistered-scope-candidate-witness")
	want := "unknown agent branch \"agent/unregistered-scope-candidate-witness\"; no paths are allowed"
	if err == nil || err.Error() != want {
		t.Fatalf("empty scope changed the existing denial: %v", err)
	}
}
