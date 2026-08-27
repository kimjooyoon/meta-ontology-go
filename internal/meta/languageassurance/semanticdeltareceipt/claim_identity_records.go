package semanticdeltareceipt

import (
	"fmt"
	"os"
	"sort"
)

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

// ClaimIdentitySourceObservation is a producer-owned observation of a real
// source pair. It deliberately keeps source evidence beside, rather than in,
// the stable claim identity.
type ClaimIdentitySourceObservation struct {
	BeforePath           string
	AfterPath            string
	BeforeRawDigest      string
	AfterRawDigest       string
	BeforeSemanticDigest string
	AfterSemanticDigest  string
	Records              []ClaimIdentityRecord
}

// ClaimIdentityObservationFromFiles reads both checked-in files and derives
// records from the canonical producer projection. The caller may compare two
// observations; neither observation is an expectation artifact.
func ClaimIdentityObservationFromFiles(input Input) (ClaimIdentitySourceObservation, error) {
	if _, err := os.Stat(input.BeforePath); err != nil {
		return ClaimIdentitySourceObservation{}, fmt.Errorf("read before source: %w", err)
	}
	if _, err := os.Stat(input.AfterPath); err != nil {
		return ClaimIdentitySourceObservation{}, fmt.Errorf("read after source: %w", err)
	}
	receipt, err := ProduceFiles(input)
	if err != nil {
		return ClaimIdentitySourceObservation{}, err
	}
	return ClaimIdentitySourceObservation{
		BeforePath: input.BeforePath, AfterPath: input.AfterPath,
		BeforeRawDigest: receipt.Before.SourceDigest, AfterRawDigest: receipt.After.SourceDigest,
		BeforeSemanticDigest: receipt.Before.SemanticDigest, AfterSemanticDigest: receipt.After.SemanticDigest,
		Records: ClaimIdentityRecords(receipt),
	}, nil
}

// ClaimIdentityPairComparison independently adjudicates persistence between
// two v3 observations. A matching claim ID is required; evidence changes do
// not create a new proposition.
type ClaimIdentityPairComparison struct {
	BaselineIDs                     []string `json:"baseline_ids"`
	AlternateIDs                    []string `json:"alternate_ids"`
	RemovedIDs                      []string `json:"removed_ids,omitempty"`
	AddedIDs                        []string `json:"added_ids,omitempty"`
	StableIdentityPreserved         int      `json:"stable_identity_preserved"`
	StableIdentityTotal             int      `json:"stable_identity_total"`
	EvidenceOnlyChanges             int      `json:"evidence_only_changes"`
	EvidenceOnlyTotal               int      `json:"evidence_only_total"`
	RawEvidenceChanged              int      `json:"raw_evidence_changed"`
	RawEvidenceTotal                int      `json:"raw_evidence_total"`
	SemanticTargetPreserved         int      `json:"semantic_target_preserved"`
	SemanticTargetTotal             int      `json:"semantic_target_total"`
	ClaimRecreatedDueOnlyToRaw      int      `json:"claim_recreated_due_only_to_raw_digest"`
	ClaimRecreatedDueOnlyToRawTotal int      `json:"claim_recreated_due_only_to_raw_digest_total"`
	Decision                        string   `json:"decision"`
	Resolution                      string   `json:"resolution"`
	Stage                           string   `json:"stage"`
	Step                            string   `json:"step"`
	Reason                          string   `json:"reason"`
}

func CompareClaimIdentityRecords(baseline, alternate []ClaimIdentityRecord) ClaimIdentityPairComparison {
	result := ClaimIdentityPairComparison{BaselineIDs: recordIDs(baseline), AlternateIDs: recordIDs(alternate), StableIdentityTotal: len(baseline), EvidenceOnlyTotal: len(baseline), RawEvidenceTotal: len(baseline), SemanticTargetTotal: len(baseline), ClaimRecreatedDueOnlyToRawTotal: len(baseline), Decision: DecisionFailClosed, Resolution: ResolutionLower, Stage: "claim-identity-persistence", Step: "compare-v3-observations", Reason: "CLAIM_IDENTITY_PERSISTENCE_UNKNOWN"}
	if !uniqueRecords(baseline) || !uniqueRecords(alternate) {
		result.Reason = "DUPLICATE_STABLE_CLAIM_ID"
		return result
	}
	result.RemovedIDs, result.AddedIDs = identitySetDiff(result.BaselineIDs, result.AlternateIDs)
	byID := make(map[string]ClaimIdentityRecord, len(alternate))
	for _, record := range alternate {
		byID[record.StableID] = record
	}
	byStableTarget := make(map[string]ClaimIdentityRecord, len(alternate))
	for _, record := range alternate {
		byStableTarget[stableRecordKey(record)] = record
	}
	for _, before := range baseline {
		after, ok := byID[before.StableID]
		stable := ok && stableRecordEqual(before, after)
		if stable {
			result.StableIdentityPreserved++
		}
		if ok && semanticTargetEqual(before, after) {
			result.SemanticTargetPreserved++
		}
		if ok && rawDigestChanged(before, after) && semanticEvidenceEqual(before, after) && semanticTargetEqual(before, after) {
			result.EvidenceOnlyChanges++
		}
		if ok && rawDigestChanged(before, after) {
			result.RawEvidenceChanged++
		}
		if !ok {
			if _, recreated := byStableTarget[stableRecordKey(before)]; recreated {
				result.ClaimRecreatedDueOnlyToRaw++
			}
		}
	}
	if len(result.RemovedIDs) == 0 && len(result.AddedIDs) == 0 && result.StableIdentityPreserved == len(baseline) && result.StableIdentityPreserved == len(alternate) {
		result.Decision, result.Resolution, result.Reason = DecisionFixedPoint, ResolutionExact, "V3_CLAIM_IDENTITY_PERSISTED_ACROSS_RAW_INTERVENTION"
	}
	return result
}

func recordIDs(records []ClaimIdentityRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.StableID)
	}
	sort.Strings(ids)
	return ids
}

func uniqueRecords(records []ClaimIdentityRecord) bool {
	seen := make(map[string]bool, len(records))
	for _, record := range records {
		if record.StableID == "" || seen[record.StableID] {
			return false
		}
		seen[record.StableID] = true
	}
	return true
}

func stableRecordKey(record ClaimIdentityRecord) string {
	return record.Kind + "\x00" + record.RelationRole + "\x00" + record.NormalizedProposition + "\x00" + record.PropositionDigest + "\x00" + record.TargetAddress + "\x00" + record.TargetAddressDigest + "\x00" + record.PreservationOf
}

func stableRecordEqual(left, right ClaimIdentityRecord) bool {
	return stableRecordKey(left) == stableRecordKey(right)
}

func semanticTargetEqual(left, right ClaimIdentityRecord) bool {
	return left.NormalizedProposition == right.NormalizedProposition && left.PropositionDigest == right.PropositionDigest && left.TargetAddress == right.TargetAddress && left.TargetAddressDigest == right.TargetAddressDigest
}

func rawDigestChanged(left, right ClaimIdentityRecord) bool {
	return left.EvidenceBeforeRawDigest != right.EvidenceBeforeRawDigest || left.EvidenceAfterRawDigest != right.EvidenceAfterRawDigest
}

func semanticEvidenceEqual(left, right ClaimIdentityRecord) bool {
	return left.EvidenceBeforeSemanticDigest == right.EvidenceBeforeSemanticDigest && left.EvidenceAfterSemanticDigest == right.EvidenceAfterSemanticDigest
}

func identitySetDiff(left, right []string) ([]string, []string) {
	l, r := map[string]bool{}, map[string]bool{}
	for _, id := range left {
		l[id] = true
	}
	for _, id := range right {
		r[id] = true
	}
	removed, added := []string{}, []string{}
	for id := range l {
		if !r[id] {
			removed = append(removed, id)
		}
	}
	for id := range r {
		if !l[id] {
			added = append(added, id)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	return removed, added
}
