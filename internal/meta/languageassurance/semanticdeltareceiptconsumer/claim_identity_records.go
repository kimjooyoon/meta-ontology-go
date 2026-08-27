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
