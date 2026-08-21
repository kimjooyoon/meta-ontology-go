package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const reconciliationLedgerPath = ".github/main-history-reconciliation.json"

type reconciliationCommit struct {
	SHA            string                      `json:"sha"`
	Message        string                      `json:"message"`
	ResultingGoMod reconciliationGoModEvidence `json:"resulting_go_mod"`
	Transition     string                      `json:"transition"`
}
type reconciliationGoModEvidence struct {
	Path        string `json:"path"`
	BlobSHA     string `json:"blob_sha"`
	Size        int    `json:"size"`
	SHA256      string `json:"sha256"`
	GoDirective string `json:"go_directive"`
}
type reconciliationBlob struct {
	Path            string `json:"path"`
	DevBlobSHA      string `json:"dev_blob_sha"`
	DevSize         int    `json:"dev_size"`
	MainBlobSHA     string `json:"main_blob_sha"`
	MainSize        int    `json:"main_size"`
	DevSHA256       string `json:"dev_sha256"`
	MainSHA256      string `json:"main_sha256"`
	DevGoDirective  string `json:"dev_go_directive"`
	MainGoDirective string `json:"main_go_directive"`
}
type reconciliationTarget struct {
	MainIsAncestorOfCandidate bool   `json:"main_is_ancestor_of_candidate"`
	FutureMainPromotion       string `json:"future_main_promotion"`
}
type reconciliationLedger struct {
	Schema                   string                      `json:"schema"`
	Version                  int                         `json:"version"`
	CandidateMergeSHA        string                      `json:"candidate_merge_sha"`
	CandidateFirstParentSHA  string                      `json:"candidate_first_parent_sha"`
	CandidateSecondParentSHA string                      `json:"candidate_second_parent_sha"`
	DevSHA                   string                      `json:"dev_sha"`
	MainSHA                  string                      `json:"main_sha"`
	MergeBaseSHA             string                      `json:"merge_base_sha"`
	MainOnlyCommits          []reconciliationCommit      `json:"main_only_commits"`
	MainOnlyPaths            []string                    `json:"main_only_paths"`
	MergeBaseBlob            reconciliationGoModEvidence `json:"merge_base_blob"`
	BlobEquivalence          reconciliationBlob          `json:"blob_equivalence"`
	EquivalenceBasis         string                      `json:"equivalence_basis"`
	Disposition              string                      `json:"disposition"`
	UnincorporatedMaterial   bool                        `json:"unincorporated_material"`
	TargetInvariant          reconciliationTarget        `json:"target_invariant"`
	CanonicalDigest          string                      `json:"canonical_digest"`
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
