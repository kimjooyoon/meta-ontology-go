package coupling

import (
	"sort"
)

func baselineChanged(changes []CodeChange, symbols map[string]CodeBinding) ([]string, bool) {
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, change := range changes {
		if !baselineID(change.CodeSymbolID) || !baselineDigest(change.BeforeDigest) || !baselineDigest(change.AfterDigest) || seen[change.CodeSymbolID] {
			return nil, false
		}
		seen[change.CodeSymbolID] = true
		if change.BeforeDigest == change.AfterDigest {
			continue
		}
		binding, exists := symbols[change.CodeSymbolID]
		if !exists {
			return nil, false
		}
		result = append(result, binding.RegisteredSurfaceID)
	}
	sort.Strings(result)
	return result, true
}
func baselineReceipts(input Input, registry map[string]CodeBinding, changed []string, before, after string) (bool, Reason) {
	wanted := map[string]bool{}
	for _, surface := range changed {
		wanted[surface] = true
	}
	seen := map[string]bool{}
	for _, receipt := range input.Receipts {
		if !baselineID(receipt.ReceiptID) {
			return false, ReasonStaleReceipt
		}
		if receipt.State == "STALE" {
			return false, ReasonStaleReceipt
		}
		if receipt.State != "CURRENT" || !wanted[receipt.SurfaceID] || seen[receipt.SurfaceID] {
			if seen[receipt.SurfaceID] {
				return false, ReasonDuplicateReceipt
			}
			if !wanted[receipt.SurfaceID] {
				return false, ReasonOrphanReceipt
			}
			return false, ReasonStaleReceipt
		}
		seen[receipt.SurfaceID] = true
		binding := registry[receipt.SurfaceID]
		if receipt.SemanticOwnerID != binding.SemanticOwnerID || receipt.CodeSymbolID != binding.CodeSymbolID || receipt.SourceMapBindingDigest != binding.BindingDigest {
			return false, ReasonRegistryBinding
		}
		if receipt.RegistryDigest != input.RegistryDigest || receipt.ToolchainDigest != input.Config.ToolchainDigest || receipt.ProfileDigest != input.Config.Profile.Digest || receipt.BeforeIRDigest != before || receipt.AfterIRDigest != after || receipt.SnapshotDigest != baselineSnapshot(input, before, after) || receipt.AuthoritySourceBeforeDigest != baselineHash(input.AuthoritySourceBefore) || receipt.AuthoritySourceAfterDigest != baselineHash(input.AuthoritySourceAfter) {
			return false, ReasonStaleReceipt
		}
	}
	if len(seen) != len(wanted) {
		return false, ReasonMissingReceipt
	}
	return len(input.Receipts) == len(wanted), ReasonOrphanReceipt
}
