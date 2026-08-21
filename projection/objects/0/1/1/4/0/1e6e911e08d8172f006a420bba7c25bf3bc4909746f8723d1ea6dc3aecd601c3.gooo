package main

import (
	"strings"
	"testing"
)

func TestMainHistoryReconciliationLedgerIsCanonical(t *testing.T) {
	ledger := readReconciliationLedger(t)
	if err := validateReconciliationLedger(ledger); err != nil {
		t.Fatal(err)
	}
}
func TestMainHistoryReconciliationRejectsTamperAndTopologyDrift(t *testing.T) {
	base := readReconciliationLedger(t)
	mutations := map[string]func(*reconciliationLedger){
		"digest tamper": func(candidate *reconciliationLedger) {
			candidate.CanonicalDigest = "sha256:" + strings.Repeat("0", 64)
		},
		"unknown residual path": func(candidate *reconciliationLedger) {
			candidate.MainOnlyPaths = append(candidate.MainOnlyPaths, "scripts/ci-proof/unknown.go")
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"unequal go.mod blobs": func(candidate *reconciliationLedger) {
			candidate.BlobEquivalence.MainSHA256 = strings.Repeat("0", 64)
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"malformed topology": func(candidate *reconciliationLedger) {
			candidate.CandidateSecondParentSHA = candidate.DevSHA
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"intermediate blob tamper": func(candidate *reconciliationLedger) {
			candidate.MainOnlyCommits[0].ResultingGoMod.BlobSHA = candidate.MainOnlyCommits[1].ResultingGoMod.BlobSHA
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"intermediate digest tamper": func(candidate *reconciliationLedger) {
			candidate.MainOnlyCommits[0].ResultingGoMod.SHA256 = strings.Repeat("0", 64)
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"intermediate directive tamper": func(candidate *reconciliationLedger) {
			candidate.MainOnlyCommits[0].ResultingGoMod.GoDirective = "go 1.26.5"
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"intermediate order tamper": func(candidate *reconciliationLedger) {
			candidate.MainOnlyCommits[0], candidate.MainOnlyCommits[1] = candidate.MainOnlyCommits[1], candidate.MainOnlyCommits[0]
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
		"intermediate falsely equals dev": func(candidate *reconciliationLedger) {
			candidate.MainOnlyCommits[0].ResultingGoMod.BlobSHA = candidate.BlobEquivalence.DevBlobSHA
			candidate.MainOnlyCommits[0].ResultingGoMod.Size = candidate.BlobEquivalence.DevSize
			candidate.MainOnlyCommits[0].ResultingGoMod.SHA256 = candidate.BlobEquivalence.DevSHA256
			candidate.MainOnlyCommits[0].ResultingGoMod.GoDirective = candidate.BlobEquivalence.DevGoDirective
			candidate.CanonicalDigest = reconciliationCanonicalDigest(*candidate)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			candidate := base
			candidate.MainOnlyPaths = append([]string(nil), base.MainOnlyPaths...)
			candidate.MainOnlyCommits = append([]reconciliationCommit(nil), base.MainOnlyCommits...)
			mutate(&candidate)
			if err := validateReconciliationLedger(candidate); err == nil {
				t.Fatal("tampered reconciliation ledger was accepted")
			}
		})
	}
}
