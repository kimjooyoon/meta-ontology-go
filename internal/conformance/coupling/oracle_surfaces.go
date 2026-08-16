package coupling

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func inputObservationCounts(input Input) ObservationCounts {
	counts := ObservationCounts{
		RegistryBindings: uint64(len(input.Registry)), ReceiptRecords: uint64(len(input.Receipts)),
		PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence)),
	}
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			counts.ChangedCodeSurfaces++
		}
	}
	for _, edge := range input.Path.Edges {
		switch edge.Kind {
		case semantic.InferenceObservationCandidate:
			counts.CandidateObservations++
		case semantic.InferenceAcceptedLift:
			counts.AcceptedLifts++
		}
	}
	return counts
}

func normalizeRegistry(bindings []CodeBinding) (registryView, oracleValidation) {
	view := registryView{bySurface: make(map[string]CodeBinding), bySymbol: make(map[string]CodeBinding)}
	canonical := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if !validID(binding.RegisteredSurfaceID) || !validID(binding.CodeSymbolID) || !validID(binding.SemanticOwnerID) ||
			!validToken(binding.SourceMapID) || !validDigest(binding.BindingDigest) {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if _, exists := view.bySurface[binding.RegisteredSurfaceID]; exists {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if _, exists := view.bySymbol[binding.CodeSymbolID]; exists {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if expected := bindingDigest(binding); expected != binding.BindingDigest {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		view.bySurface[binding.RegisteredSurfaceID] = binding
		view.bySymbol[binding.CodeSymbolID] = binding
		canonical = append(canonical, bindingCanonical(binding))
	}
	sort.Strings(canonical)
	view.digest = digestBytes([]byte(strings.Join(canonical, "\n") + "\n"))
	return view, oracleValidation{}
}

func resolveChangedSurfaces(changes []CodeChange, registry registryView) ([]string, oracleValidation) {
	seen := make(map[string]struct{}, len(changes))
	changed := make([]string, 0, len(changes))
	for _, change := range changes {
		if !validID(change.CodeSymbolID) || !validDigest(change.BeforeDigest) || !validDigest(change.AfterDigest) {
			return nil, oracleValidation{DecisionFailClosed, ReasonChangedSurface}
		}
		if _, duplicate := seen[change.CodeSymbolID]; duplicate {
			return nil, oracleValidation{DecisionFailClosed, ReasonChangedSurface}
		}
		seen[change.CodeSymbolID] = struct{}{}
		if change.BeforeDigest == change.AfterDigest {
			continue
		}
		binding, exists := registry.bySymbol[change.CodeSymbolID]
		if !exists {
			return nil, oracleValidation{DecisionFailClosed, ReasonSurfaceUnregistered}
		}
		changed = append(changed, binding.RegisteredSurfaceID)
	}
	sort.Strings(changed)
	return changed, oracleValidation{}
}

func validateSourceBindings(input Input, beforeDigest, afterDigest string) oracleValidation {
	if sourceDigest(input.AuthoritySourceBefore) != beforeDigest && input.AuthoritySourceBefore == "" {
		return oracleValidation{DecisionUnknown, ReasonSourceUnbound}
	}
	if !validDigest(sourceDigest(input.AuthoritySourceBefore)) || !validDigest(sourceDigest(input.AuthoritySourceAfter)) {
		return oracleValidation{DecisionUnknown, ReasonSourceUnbound}
	}
	return oracleValidation{}
}

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

func validateReceiptClaim(receipt CouplingReceipt, beforeDigest, afterDigest, deltaText string) oracleValidation {
	switch receipt.ChangeClaim {
	case ClaimDelta:
		if receipt.ReceiptKind != ReceiptSemanticDelta || beforeDigest == afterDigest {
			return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
		}
		if receipt.SemanticDelta != deltaText || receipt.SemanticDelta == "" || receipt.SemanticDeltaDigest != digestBytes([]byte(deltaText)) {
			return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
		}
		if receipt.AuthoritativeSourceRef == "" || receipt.AuthoritySourceAfterDigest == "" {
			return oracleValidation{DecisionFailClosed, ReasonDeltaWithoutSource}
		}
	case ClaimNoDelta:
		if receipt.ReceiptKind != ReceiptNoSemanticDelta || beforeDigest != afterDigest {
			return oracleValidation{DecisionFailClosed, ReasonNoDeltaWithoutEquality}
		}
		if receipt.SemanticDelta != "" || receipt.SemanticDeltaDigest != "" || receipt.AuthoritativeSourceRef != "" {
			return oracleValidation{DecisionFailClosed, ReasonNoDeltaWithoutEquality}
		}
	default:
		return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
	}
	if len(receipt.EvidenceRefs) == 0 || len(receipt.OriginPathIDs) == 0 || receipt.ClaimRecordID == "" {
		return oracleValidation{DecisionFailClosed, ReasonPathMalformed}
	}
	return oracleValidation{}
}
