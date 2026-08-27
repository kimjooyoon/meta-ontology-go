package semanticdeltareceiptconsumer

import "testing"

func TestClaimIdentityRematchIncludesPreservationOf(t *testing.T) {
	baseline := claimIdentityRecordForMatchingTest("before", "preserves-old")
	alternate := claimIdentityRecordForMatchingTest("after", "preserves-new")

	comparison := CompareClaimIdentityRecords([]ClaimIdentityRecord{baseline}, []ClaimIdentityRecord{alternate})
	if comparison.Decision != decisionFailClosed || comparison.Resolution != resolutionLower || comparison.Reason != "CLAIM_SET_CHANGED" {
		t.Fatalf("preservation relation drift was misclassified: %+v", comparison)
	}
	if len(comparison.RemovedIDs) != 1 || len(comparison.AddedIDs) != 1 || comparison.ClaimRecreatedDueOnlyToRaw != 0 {
		t.Fatalf("preservation relation drift was treated as raw-only recreation: %+v", comparison)
	}
}

func TestClaimIdentityRematchRejectsAmbiguousSemanticIdentity(t *testing.T) {
	baseline := claimIdentityRecordForMatchingTest("before", "preserves")
	first := claimIdentityRecordForMatchingTest("after-1", "preserves")
	second := claimIdentityRecordForMatchingTest("after-2", "preserves")

	comparison := CompareClaimIdentityRecords([]ClaimIdentityRecord{baseline}, []ClaimIdentityRecord{first, second})
	if comparison.Decision != decisionFailClosed || comparison.Resolution != resolutionLower || comparison.Reason != "AMBIGUOUS_CLAIM_IDENTITY_MATCH" {
		t.Fatalf("ambiguous semantic identity was not fail-closed: %+v", comparison)
	}
}

func claimIdentityRecordForMatchingTest(stableID, preservationOf string) ClaimIdentityRecord {
	return ClaimIdentityRecord{
		StableID:                     stableID,
		Kind:                         claimKindPreserve,
		RelationRole:                 "preserves",
		NormalizedProposition:        "preserve\x00order\x00uses\x00payment",
		PropositionDigest:            "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		TargetAddress:                "order\x00uses\x00payment",
		TargetAddressDigest:          "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		PreservationOf:               preservationOf,
		EvidenceBeforeRawDigest:      "sha256:3333333333333333333333333333333333333333333333333333333333333333",
		EvidenceAfterRawDigest:       "sha256:4444444444444444444444444444444444444444444444444444444444444444",
		EvidenceBeforeSemanticDigest: "sha256:5555555555555555555555555555555555555555555555555555555555555555",
		EvidenceAfterSemanticDigest:  "sha256:6666666666666666666666666666666666666666666666666666666666666666",
	}
}
