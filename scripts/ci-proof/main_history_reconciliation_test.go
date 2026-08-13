package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const reconciliationLedgerPath = ".github/main-history-reconciliation.json"

type reconciliationCommit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
}

type reconciliationBlob struct {
	Path        string `json:"path"`
	DevBlobSHA  string `json:"dev_blob_sha"`
	MainBlobSHA string `json:"main_blob_sha"`
	DevSHA256   string `json:"dev_sha256"`
	MainSHA256  string `json:"main_sha256"`
}

type reconciliationTarget struct {
	MainIsAncestorOfCandidate bool   `json:"main_is_ancestor_of_candidate"`
	FutureMainPromotion       string `json:"future_main_promotion"`
}

type reconciliationLedger struct {
	Schema                   string                 `json:"schema"`
	Version                  int                    `json:"version"`
	CandidateMergeSHA        string                 `json:"candidate_merge_sha"`
	CandidateFirstParentSHA  string                 `json:"candidate_first_parent_sha"`
	CandidateSecondParentSHA string                 `json:"candidate_second_parent_sha"`
	DevSHA                   string                 `json:"dev_sha"`
	MainSHA                  string                 `json:"main_sha"`
	MergeBaseSHA             string                 `json:"merge_base_sha"`
	MainOnlyCommits          []reconciliationCommit `json:"main_only_commits"`
	MainOnlyPaths            []string               `json:"main_only_paths"`
	BlobEquivalence          reconciliationBlob     `json:"blob_equivalence"`
	Disposition              string                 `json:"disposition"`
	UnincorporatedMaterial   bool                   `json:"unincorporated_material"`
	TargetInvariant          reconciliationTarget   `json:"target_invariant"`
	CanonicalDigest          string                 `json:"canonical_digest"`
}

func readReconciliationLedger(t *testing.T) reconciliationLedger {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", reconciliationLedgerPath))
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var ledger reconciliationLedger
	if err := decoder.Decode(&ledger); err != nil {
		t.Fatal(err)
	}
	return ledger
}

func reconciliationCanonicalDigest(ledger reconciliationLedger) string {
	ledger.CanonicalDigest = ""
	data, err := json.Marshal(ledger)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validReconciliationSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, char := range value {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return false
		}
	}
	return true
}

func validReconciliationDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(value) == len("sha256:")+64
}

func validateReconciliationLedger(ledger reconciliationLedger) error {
	if ledger.Schema != "gooo/main-history-reconciliation/v1" || ledger.Version != 1 || !validReconciliationSHA(ledger.CandidateMergeSHA) || !validReconciliationSHA(ledger.CandidateFirstParentSHA) || !validReconciliationSHA(ledger.CandidateSecondParentSHA) || !validReconciliationSHA(ledger.DevSHA) || !validReconciliationSHA(ledger.MainSHA) || !validReconciliationSHA(ledger.MergeBaseSHA) {
		return fmt.Errorf("reconciliation schema or topology SHA is malformed")
	}
	if ledger.CandidateMergeSHA != "6a95e73a391a114d8a22cdc23251e8a7623d7768" || ledger.DevSHA != "3b1647c7aa14558d07f868d649071a2c36e6cc88" || ledger.MainSHA != "c557daf1fd6748b2e61afca10fb632792683061f" || ledger.MergeBaseSHA != "0e838bff165dc73ed984ecbdca338d957b8abfa9" {
		return fmt.Errorf("reconciliation live topology SHA evidence is not exact")
	}
	if ledger.CandidateFirstParentSHA != ledger.DevSHA || ledger.CandidateSecondParentSHA != ledger.MainSHA || !ledger.TargetInvariant.MainIsAncestorOfCandidate {
		return fmt.Errorf("candidate parent topology is not exact")
	}
	if ledger.TargetInvariant.FutureMainPromotion != "exact_ci_backed_fast_forward_plus_main_dev_equality" {
		return fmt.Errorf("future promotion invariant is not exact")
	}
	if len(ledger.MainOnlyCommits) != 2 || ledger.MainOnlyCommits[0].SHA != "a798c8be290fd7f96662edca173cd93cfe270482" || ledger.MainOnlyCommits[0].Message != "chore: require Go 1.26.6" || ledger.MainOnlyCommits[1].SHA != "c557daf1fd6748b2e61afca10fb632792683061f" || ledger.MainOnlyCommits[1].Message != "chore: use Go 1.26.5 toolchain" {
		return fmt.Errorf("main-only commit evidence is not exact")
	}
	if !sort.StringsAreSorted(ledger.MainOnlyPaths) || len(ledger.MainOnlyPaths) != 1 || ledger.MainOnlyPaths[0] != "go.mod" {
		return fmt.Errorf("main-only path inventory is not the exact sorted residual set")
	}
	blob := ledger.BlobEquivalence
	if blob.Path != "go.mod" || !validReconciliationSHA(blob.DevBlobSHA) || blob.DevBlobSHA != blob.MainBlobSHA || !validReconciliationDigest("sha256:"+blob.DevSHA256) || blob.DevSHA256 != blob.MainSHA256 {
		return fmt.Errorf("go.mod blob equivalence is not exact")
	}
	if ledger.Disposition != "TREE_EQUIVALENT_ALREADY_IN_DEV" || ledger.UnincorporatedMaterial {
		return fmt.Errorf("reconciliation disposition is unsafe")
	}
	if ledger.CanonicalDigest != reconciliationCanonicalDigest(ledger) {
		return fmt.Errorf("reconciliation canonical digest does not match")
	}
	return nil
}

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
