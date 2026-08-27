package semanticdeltareceiptconsumer

import (
	"fmt"
	"os"
	"sort"
)

// ClaimIdentityRecordsFromFiles is an independent reconstruction path. It
// reads raw source and lowers it through the consumer's copied lowering path;
// producer receipt fields are never used as evidence.
func ClaimIdentityRecordsFromFiles(input Input) ([]ClaimIdentityRecord, SourcePairObservation, error) {
	beforeRaw, err := os.ReadFile(input.BeforePath)
	if err != nil {
		return nil, SourcePairObservation{}, fmt.Errorf("read before source: %w", err)
	}
	afterRaw, err := os.ReadFile(input.AfterPath)
	if err != nil {
		return nil, SourcePairObservation{}, fmt.Errorf("read after source: %w", err)
	}
	receipt := reconstructReceipt(input, beforeRaw, afterRaw)
	if receipt.ClaimTransitionIdentityDigest == "" || receipt.ClaimIDInventory == nil {
		return nil, SourcePairObservation{}, fmt.Errorf("consumer reconstruction did not produce identity evidence")
	}
	claims := claimIdentityRecords(receipt.ClaimLedger)
	sort.Slice(claims, func(i, j int) bool { return claims[i].StableID < claims[j].StableID })
	return claims, SourcePairObservation{BeforePath: input.BeforePath, AfterPath: input.AfterPath, BeforeRawDigest: receipt.Before.SourceDigest, AfterRawDigest: receipt.After.SourceDigest, BeforeSemanticDigest: receipt.Before.SemanticDigest, AfterSemanticDigest: receipt.After.SemanticDigest}, nil
}

// ClaimIdentityPairComparison is independently computed by the consumer from
// two raw-source observations. It is deliberately a copied wire model rather
// than a producer result.
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

// CompareClaimIdentityRecords matches the same observation slot across two
// receipts. Raw paths/digests are evidence; stable proposition fields define
// the persistent claim identity.
func CompareClaimIdentityRecords(baseline, alternate []ClaimIdentityRecord) ClaimIdentityPairComparison {
	result := ClaimIdentityPairComparison{BaselineIDs: recordIDs(baseline), AlternateIDs: recordIDs(alternate), StableIdentityTotal: len(baseline), EvidenceOnlyTotal: len(baseline), RawEvidenceTotal: len(baseline), SemanticTargetTotal: len(baseline), ClaimRecreatedDueOnlyToRawTotal: len(baseline), Decision: decisionFailClosed, Resolution: resolutionLower, Stage: "claim-identity-persistence", Step: "compare-v3-observations", Reason: "CLAIM_IDENTITY_PERSISTENCE_UNKNOWN"}
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
		if ok && stableRecordEqual(before, after) {
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
		result.Decision, result.Resolution, result.Reason = decisionFixedPoint, resolutionExact, "V3_CLAIM_IDENTITY_PERSISTED_ACROSS_RAW_INTERVENTION"
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
