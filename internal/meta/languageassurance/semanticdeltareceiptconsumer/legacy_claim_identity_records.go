package semanticdeltareceiptconsumer

import (
	"fmt"
	"os"
	"strings"
)

// LegacyClaimIdentityRecordsFromFiles independently reconstructs the v2
// evidence-bound identity used by the preserved old expectation artifact. It
// is intentionally separate from the v3 identity path so evolution can prove
// that both producer and consumer observed the same old inventory before
// comparing it with the new stable identity inventory.
func LegacyClaimIdentityRecordsFromFiles(input Input) ([]ClaimIdentityRecord, SourcePairObservation, error) {
	beforeRaw, err := os.ReadFile(input.BeforePath)
	if err != nil {
		return nil, SourcePairObservation{}, fmt.Errorf("read before source: %w", err)
	}
	afterRaw, err := os.ReadFile(input.AfterPath)
	if err != nil {
		return nil, SourcePairObservation{}, fmt.Errorf("read after source: %w", err)
	}
	receipt := reconstructReceipt(input, beforeRaw, afterRaw)
	before := legacyObjectRecords(receipt.Before.Claims, input.BeforePath, receipt.Before.SourceDigest, receipt.Before.SemanticDigest, "before")
	after := legacyObjectRecords(receipt.After.Claims, input.AfterPath, receipt.After.SourceDigest, receipt.After.SemanticDigest, "after")
	result := make([]ClaimIdentityRecord, 0, 1+len(before)+len(after)+len(before))
	boundedTarget := input.BeforePath + "->" + input.AfterPath
	boundedNormalized := strings.Join([]string{"BOUNDED_SEMANTIC_EQUIVALENCE", "source-pair", "bounded-semantic-equivalence", input.BeforePath + "\x00" + input.AfterPath + "\x00" + receipt.Before.SourceDigest + "\x00" + receipt.After.SourceDigest + "\x00" + receipt.Before.SemanticDigest + "\x00" + receipt.After.SemanticDigest}, "\x00")
	boundedDigest := digestValue(boundedNormalized)
	result = append(result, ClaimIdentityRecord{StableID: "gooo://semantic-delta/claim/bounded-equivalence/" + boundedDigest[len("sha256:"):], Kind: claimKindBounded, RelationRole: "bounded-equivalence", NormalizedProposition: boundedNormalized, PropositionDigest: boundedDigest, TargetAddress: boundedTarget, TargetAddressDigest: targetAddressDigest(boundedTarget), BeforeSourcePath: input.BeforePath, AfterSourcePath: input.AfterPath, EvidenceBeforeRawDigest: receipt.Before.SourceDigest, EvidenceAfterRawDigest: receipt.After.SourceDigest, EvidenceBeforeSemanticDigest: receipt.Before.SemanticDigest, EvidenceAfterSemanticDigest: receipt.After.SemanticDigest})
	result = append(result, before...)
	result = append(result, after...)
	if receipt.Classification == classPreserved || receipt.Classification == classChanged {
		for _, oldBefore := range before {
			var match ClaimIdentityRecord
			for _, candidate := range after {
				if oldBefore.NormalizedProposition == candidate.NormalizedProposition {
					match = candidate
					break
				}
			}
			afterTarget, afterRawDigest, afterSemanticDigest := "", "", ""
			if match.StableID != "" {
				afterTarget, afterRawDigest, afterSemanticDigest = match.TargetAddress, receipt.After.SourceDigest, receipt.After.SemanticDigest
			}
			identity := digestValue(strings.Join([]string{oldBefore.StableID, match.StableID, afterTarget, oldBefore.NormalizedProposition, afterRawDigest, afterSemanticDigest}, "\x00"))
			claimTypeID := digestValue(oldBefore.NormalizedProposition)
			normalized := strings.Join([]string{claimKindPreserve, claimTypeID, "preserves", oldBefore.NormalizedProposition}, "\x00")
			result = append(result, ClaimIdentityRecord{StableID: "gooo://semantic-delta/claim/preservation/" + identity[len("sha256:"):], Kind: claimKindPreserve, RelationRole: "preserves", NormalizedProposition: normalized, PropositionDigest: digestValue(normalized), TargetAddress: oldBefore.TargetAddress, TargetAddressDigest: targetAddressDigest(oldBefore.TargetAddress), PreservationOf: oldBefore.StableID, BeforeSourcePath: input.BeforePath, AfterSourcePath: match.AfterSourcePath, EvidenceBeforeRawDigest: receipt.Before.SourceDigest, EvidenceAfterRawDigest: afterRawDigest, EvidenceBeforeSemanticDigest: receipt.Before.SemanticDigest, EvidenceAfterSemanticDigest: afterSemanticDigest})
		}
	}
	return result, SourcePairObservation{BeforePath: input.BeforePath, AfterPath: input.AfterPath, BeforeRawDigest: receipt.Before.SourceDigest, AfterRawDigest: receipt.After.SourceDigest, BeforeSemanticDigest: receipt.Before.SemanticDigest, AfterSemanticDigest: receipt.After.SemanticDigest}, nil
}

func legacyObjectRecords(claims []Claim, path, rawDigest, semanticDigest, side string) []ClaimIdentityRecord {
	result := make([]ClaimIdentityRecord, 0, len(claims))
	for _, claim := range claims {
		identity := digestValue(strings.Join([]string{claim.NormalizedProposition, path, rawDigest, semanticDigest}, "\x00"))
		result = append(result, ClaimIdentityRecord{StableID: "gooo://semantic-delta/claim/object/" + identity[len("sha256:"):], Kind: claim.Kind, RelationRole: claim.Predicate + "|observation|" + side, NormalizedProposition: claim.NormalizedProposition, PropositionDigest: claim.PropositionDigest, TargetAddress: path, TargetAddressDigest: targetAddressDigest(path), BeforeSourcePath: chooseLegacyPath(path, side, "before"), AfterSourcePath: chooseLegacyPath(path, side, "after"), EvidenceBeforeRawDigest: chooseLegacyDigest(rawDigest, side, "before"), EvidenceAfterRawDigest: chooseLegacyDigest(rawDigest, side, "after"), EvidenceBeforeSemanticDigest: chooseLegacyDigest(semanticDigest, side, "before"), EvidenceAfterSemanticDigest: chooseLegacyDigest(semanticDigest, side, "after")})
	}
	return result
}

func chooseLegacyPath(path, side, wanted string) string {
	if side == wanted {
		return path
	}
	return ""
}

func chooseLegacyDigest(digest, side, wanted string) string {
	if side == wanted {
		return digest
	}
	return ""
}
