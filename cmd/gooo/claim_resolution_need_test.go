package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestClaimResolutionNeedIsDigestLocked(t *testing.T) {
	lockBytes, err := os.ReadFile("testdata/claim-resolution/need-lock.json")
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		Tag             string `json:"tag"`
		TagObjectSHA    string `json:"tag_object_sha"`
		TargetCommitSHA string `json:"target_commit_sha"`
		SHA256          string `json:"sha256"`
	}
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		t.Fatal(err)
	}
	receipt, err := os.ReadFile("testdata/claim-resolution/primitive-need.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(receipt)
	if hex.EncodeToString(sum[:]) != lock.SHA256 || lock.Tag != "v0.9.0-dev" || len(lock.TagObjectSHA) != 40 || len(lock.TargetCommitSHA) != 40 {
		t.Fatalf("need identity changed: %#v", lock)
	}
	var observed struct {
		Decision string `json:"decision"`
		Summary struct {
			Closed              int `json:"closed"`
			ConsumerReleases    int `json:"consumer_releases"`
			CommonEnvelopes     int `json:"common_envelopes"`
			ClaimStates         int `json:"claim_states"`
			UnknownProducers    int `json:"unknown_producers"`
			TypedUnknownRoles   int `json:"typed_unknown_roles"`
			CompatibilityGaps   int `json:"compatibility_gaps"`
			DirectMappings      int `json:"direct_mappings"`
		} `json:"summary"`
		Candidate struct {
			ID                   string `json:"id"`
			ImplementationStatus string `json:"implementation_status"`
		} `json:"candidate"`
		Claim struct {
			NextOperation string `json:"next_operation"`
		} `json:"claim"`
	}
	if err := json.Unmarshal(receipt, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Decision != "CROSS_CONSUMER_PRIMITIVE_NEED_OBSERVED" || observed.Summary.Closed != 12 || observed.Summary.ConsumerReleases != 3 || observed.Summary.CommonEnvelopes != 3 || observed.Summary.ClaimStates != 3 || observed.Summary.UnknownProducers != 2 || observed.Summary.TypedUnknownRoles != 2 || observed.Summary.CompatibilityGaps != 1 || observed.Summary.DirectMappings != 0 || observed.Candidate.ID != claimResolutionCandidateID || observed.Candidate.ImplementationStatus != "NOT_SELECTED" || observed.Claim.NextOperation != "DEFINE_MINIMAL_CLAIM_RESOLUTION_CORE_CONTRACT" {
		t.Fatalf("cross-consumer need changed: %#v", observed)
	}
}
