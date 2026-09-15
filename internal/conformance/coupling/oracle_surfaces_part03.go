package coupling

import (
	"sort"
)

func validateReceipts(input Input, registry registryView, changed []string, before, after normalizedSemantic, deltaText string) (receiptView, oracleValidation) {
	view := receiptView{bySurface: make(map[string]CouplingReceipt)}
	changedSet := make(map[string]struct{}, len(changed))
	for _, surface := range changed {
		changedSet[surface] = struct{}{}
	}
	seenIDs := make(map[string]struct{}, len(input.Receipts))
	for _, receipt := range input.Receipts {
		if !validID(receipt.ReceiptID) {
			return view, oracleValidation{DecisionFailClosed, ReasonStaleReceipt}
		}
		if receipt.State == "STALE" {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.State != "CURRENT" {
			return view, oracleValidation{DecisionFailClosed, ReasonStaleReceipt}
		}
		if _, duplicate := seenIDs[receipt.ReceiptID]; duplicate {
			return view, oracleValidation{DecisionFailClosed, ReasonDuplicateReceipt}
		}
		seenIDs[receipt.ReceiptID] = struct{}{}
		if _, exists := changedSet[receipt.SurfaceID]; !exists {
			return view, oracleValidation{DecisionFailClosed, ReasonOrphanReceipt}
		}
		if _, duplicate := view.bySurface[receipt.SurfaceID]; duplicate {
			return view, oracleValidation{DecisionFailClosed, ReasonDuplicateReceipt}
		}
		binding := registry.bySurface[receipt.SurfaceID]
		if receipt.SemanticOwnerID != binding.SemanticOwnerID || receipt.CodeSymbolID != binding.CodeSymbolID ||
			receipt.SourceMapBindingDigest != binding.BindingDigest {
			return view, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		expectedSnapshot := snapshotDigest(input, before.digest, after.digest, registry.digest)
		if receipt.SnapshotDigest != expectedSnapshot || receipt.RegistryDigest != registry.digest {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.ToolchainDigest != input.Config.ToolchainDigest || receipt.ProfileDigest != input.Config.Profile.Digest {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.BeforeIRDigest != before.digest || receipt.AfterIRDigest != after.digest {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if receipt.AuthoritySourceBeforeDigest != sourceDigest(input.AuthoritySourceBefore) || receipt.AuthoritySourceAfterDigest != sourceDigest(input.AuthoritySourceAfter) {
			return view, oracleValidation{DecisionUnknown, ReasonStaleReceipt}
		}
		if issue := validateReceiptClaim(receipt, before.digest, after.digest, deltaText); issue.decision != "" {
			return view, issue
		}
		view.bySurface[receipt.SurfaceID] = receipt
		view.valid = append(view.valid, receipt.SurfaceID)
	}
	if len(view.valid) != len(changed) {
		for _, surface := range changed {
			if _, exists := view.bySurface[surface]; !exists {
				return view, oracleValidation{DecisionUnknown, ReasonMissingReceipt}
			}
		}
		return view, oracleValidation{DecisionUnknown, ReasonMissingReceipt}
	}
	sort.Strings(view.valid)
	return view, oracleValidation{}
}
