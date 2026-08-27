package semanticdeltareceipt

import "sort"

// ClaimIdentityRecord is the producer's copied wire description of a
// persistent claim. The stable fields identify the proposition; the evidence
// fields identify the observation that currently supports it.
type ClaimIdentityRecord struct {
	StableID                     string `json:"stable_id"`
	Kind                         string `json:"kind"`
	RelationRole                 string `json:"relation_role"`
	NormalizedProposition        string `json:"normalized_proposition"`
	PropositionDigest            string `json:"proposition_digest"`
	TargetAddress                string `json:"target_address"`
	TargetAddressDigest          string `json:"target_address_digest"`
	PreservationOf               string `json:"preservation_of,omitempty"`
	BeforeSourcePath             string `json:"before_source_path,omitempty"`
	AfterSourcePath              string `json:"after_source_path,omitempty"`
	EvidenceBeforeRawDigest      string `json:"evidence_before_raw_digest,omitempty"`
	EvidenceAfterRawDigest       string `json:"evidence_after_raw_digest,omitempty"`
	EvidenceBeforeSemanticDigest string `json:"evidence_before_semantic_digest,omitempty"`
	EvidenceAfterSemanticDigest  string `json:"evidence_after_semantic_digest,omitempty"`
}

func ClaimIdentityRecords(receipt Receipt) []ClaimIdentityRecord {
	result := make([]ClaimIdentityRecord, 0, len(receipt.ClaimLedger))
	for _, claim := range receipt.ClaimLedger {
		result = append(result, ClaimIdentityRecord{StableID: claim.ID, Kind: claim.Kind, RelationRole: claim.RelationRole, NormalizedProposition: claim.NormalizedProposition, PropositionDigest: claim.PropositionDigest, TargetAddress: claim.TargetAddress, TargetAddressDigest: claim.TargetAddressDigest, PreservationOf: claim.PreservationOf, BeforeSourcePath: claim.BeforeSourcePath, AfterSourcePath: claim.AfterSourcePath, EvidenceBeforeRawDigest: claim.BeforeSourceDigest, EvidenceAfterRawDigest: claim.AfterSourceDigest, EvidenceBeforeSemanticDigest: claim.BeforeSemanticDigest, EvidenceAfterSemanticDigest: claim.AfterSemanticDigest})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StableID < result[j].StableID })
	return result
}

// ClaimIdentityRecordsFromFiles deliberately goes through the producer's
// source-reading boundary; consumers must not use this function to adjudicate.
func ClaimIdentityRecordsFromFiles(input Input) ([]ClaimIdentityRecord, error) {
	receipt, err := ProduceFiles(input)
	if err != nil {
		return nil, err
	}
	return ClaimIdentityRecords(receipt), nil
}
