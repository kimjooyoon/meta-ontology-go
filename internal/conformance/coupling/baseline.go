package coupling

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// EvaluateBaseline is a deliberately separate full-suite implementation. It
// re-derives registry closure, semantic equality, receipt predicates, and
// resource bindings from the typed input without calling Evaluate or any of
// the oracle's private helpers.
func EvaluateBaseline(input Input) BaselineResult {
	result := BaselineResult{FullSuite: true, ObservationCounts: baselineCounts(input)}
	if !baselineResourceBindingsEqual(input.Config.ResourceBinding, input.ResourceRegistry) {
		return baselineUnknown(result, input, ReasonResourceUnbound)
	}
	resources, ok, reason := baselineResources(input.ResourceReceipts, input.Config.ResourceBinding)
	result.Resources = resources
	result.WorkUnits = resources.WorkUnits
	if !ok {
		return baselineUnknown(result, input, reason)
	}
	if input.Schema != SchemaV1 || input.Config.ToolchainDigest == "" || input.Config.Profile.ID == "" || input.Config.Profile.Version == "" || input.Config.Profile.Digest == "" || input.AuthoritySourceBefore == "" || input.AuthoritySourceAfter == "" || len(input.Registry) == 0 || !input.Manifest.Complete || (len(input.Changes) == 0 && !input.Manifest.ZeroChange) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	if !baselineDigest(input.Config.ToolchainDigest) || !baselineDigest(input.Config.Profile.Digest) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	before, ok := baselineSemantic(input.SemanticBefore)
	if !ok {
		return baselineFail(result, input, ReasonPathMalformed)
	}
	after, ok := baselineSemantic(input.SemanticAfter)
	if !ok {
		return baselineFail(result, input, ReasonPathMalformed)
	}
	registry, ok := baselineRegistry(input.Registry)
	if !ok || registry.digest != input.RegistryDigest {
		return baselineFail(result, input, ReasonDigestMismatch)
	}
	if !baselineManifest(input, before.digest, after.digest, registry.digest) {
		if !input.Manifest.Complete {
			return baselineUnknown(result, input, ReasonRequiredInputMissing)
		}
		return baselineFail(result, input, ReasonDigestMismatch)
	}
	changed, ok := baselineChanged(input.Changes, registry.bySymbol)
	if !ok {
		return baselineFail(result, input, ReasonChangedSurface)
	}
	result.LocalizedSurfaces = append([]string(nil), changed...)
	if receiptsOK, receiptReason := baselineReceipts(input, registry.bySurface, changed, before.digest, after.digest); !receiptsOK {
		if receiptReason == ReasonMissingReceipt || receiptReason == ReasonStaleReceipt {
			return baselineUnknown(result, input, receiptReason)
		}
		return baselineFail(result, input, receiptReason)
	}
	delta := baselineDelta(before.facts, after.facts)
	if !baselineClaims(input.Receipts, before.digest, after.digest, delta) {
		return baselineFail(result, input, ReasonInvalidDelta)
	}
	if len(changed) > 0 && !baselinePath(input, registry.bySurface, input.Receipts, before.digest, after.digest, delta) {
		return baselineFail(result, input, ReasonPathClosure)
	}
	result.Decision, result.Reason, result.LocalizedSurfaces = DecisionPass, ReasonNone, nil
	return result
}

func baselineResourceBindingsEqual(left, right ResourceBindingConfig) bool {
	return left.ProviderID != "" && left.ProviderID == right.ProviderID && left.ObserverID != "" && left.ObserverID == right.ObserverID && left.ProviderDigest != "" && left.ProviderDigest == right.ProviderDigest && left.ObserverDigest != "" && left.ObserverDigest == right.ObserverDigest && left.SnapshotDigest != "" && left.SnapshotDigest == right.SnapshotDigest && left.SourceDigest != "" && left.SourceDigest == right.SourceDigest
}

func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	baseline := EvaluateBaseline(input)
	comparison := Comparison{Oracle: oracle, Baseline: baseline, OutcomeMatch: oracle.Decision == baseline.Decision, ReasonMatch: oracle.Reason == baseline.Reason, LocalizationMatch: sameSurfaceSet(oracle.ChangedSurfaces, baseline.LocalizedSurfaces)}
	if oracle.Decision == DecisionPass {
		comparison.LocalizationMatch = len(baseline.LocalizedSurfaces) == 0
	}
	if comparison.OutcomeMatch && comparison.ReasonMatch && comparison.LocalizationMatch {
		comparison.Finding = "NO_UNIQUE_BENEFIT"
	} else {
		comparison.Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
	}
	return comparison
}

type baselineSemanticView struct {
	digest string
	facts  []string
}

type baselineRegistryView struct {
	bySurface map[string]CodeBinding
	bySymbol  map[string]CodeBinding
	digest    string
}

func baselineCounts(input Input) ObservationCounts {
	counts := ObservationCounts{RegistryBindings: uint64(len(input.Registry)), ReceiptRecords: uint64(len(input.Receipts)), PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence)), ResourceReceipts: uint64(len(input.ResourceReceipts))}
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			counts.ChangedCodeSurfaces++
		}
	}
	for _, edge := range input.Path.Edges {
		if edge.Kind == semantic.InferenceObservationCandidate {
			counts.CandidateObservations++
		}
		if edge.Kind == semantic.InferenceAcceptedLift {
			counts.AcceptedLifts++
		}
	}
	return counts
}

func baselineResources(receipts []ExternalResourceReceipt, expected ResourceBindingConfig) (ResourceObservation, bool, Reason) {
	seen := map[string]bool{}
	var result ResourceObservation
	for _, receipt := range receipts {
		if !baselineID(receipt.ReceiptID) || !baselineID(expected.ProviderID) || !baselineID(expected.ObserverID) || !baselineDigest(expected.ProviderDigest) || !baselineDigest(expected.ObserverDigest) || !baselineDigest(expected.SnapshotDigest) || !baselineDigest(expected.SourceDigest) || !baselineDigest(receipt.ProviderDigest) || !baselineDigest(receipt.ObserverDigest) || !baselineDigest(receipt.SnapshotDigest) || !baselineDigest(receipt.SourceDigest) || !baselineDigest(receipt.BindingDigest) || expected.ProviderDigest != baselineProviderDigest(expected.ProviderID) || expected.ObserverDigest != baselineObserverDigest(expected.ObserverID) || expected.SourceDigest != baselineSourceDigest(expected.ProviderID, expected.ObserverID, expected.SnapshotDigest) || receipt.ProviderDigest != expected.ProviderDigest || receipt.ObserverDigest != expected.ObserverDigest || receipt.SnapshotDigest != expected.SnapshotDigest || receipt.SourceDigest != expected.SourceDigest || !receipt.Present || !receipt.Independent || receipt.State != "CURRENT" || receipt.BindingDigest != baselineResourceDigest(receipt) {
			return ResourceObservation{}, false, ReasonResourceUnbound
		}
		if seen[receipt.Metric] {
			return ResourceObservation{}, false, ReasonResourceUnbound
		}
		seen[receipt.Metric] = true
		switch receipt.Metric {
		case "cpu-core-ns":
			if receipt.Unit != "ns" {
				return ResourceObservation{}, false, ReasonResourceUnbound
			}
			result.CPUCoreNS = receipt.Value
		case "peak-memory-bytes":
			if receipt.Unit != "bytes" {
				return ResourceObservation{}, false, ReasonResourceUnbound
			}
			result.PeakMemoryBytes = receipt.Value
		case "work-units":
			if receipt.Unit != "units" {
				return ResourceObservation{}, false, ReasonResourceUnbound
			}
			result.WorkUnits = receipt.Value
		default:
			return ResourceObservation{}, false, ReasonResourceUnbound
		}
	}
	if len(seen) != 3 {
		return ResourceObservation{}, false, ReasonResourceUnbound
	}
	return result, true, ReasonNone
}

func baselineRegistry(bindings []CodeBinding) (baselineRegistryView, bool) {
	result := baselineRegistryView{bySurface: map[string]CodeBinding{}, bySymbol: map[string]CodeBinding{}}
	parts := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if !baselineID(binding.RegisteredSurfaceID) || !baselineID(binding.CodeSymbolID) || !baselineID(binding.SemanticOwnerID) || binding.SourceMapID == "" || !baselineDigest(binding.BindingDigest) || binding.BindingDigest != baselineBindingDigest(binding) {
			return baselineRegistryView{}, false
		}
		if _, exists := result.bySurface[binding.RegisteredSurfaceID]; exists {
			return baselineRegistryView{}, false
		}
		if _, exists := result.bySymbol[binding.CodeSymbolID]; exists {
			return baselineRegistryView{}, false
		}
		result.bySurface[binding.RegisteredSurfaceID] = binding
		result.bySymbol[binding.CodeSymbolID] = binding
		parts = append(parts, baselineBindingCanonical(binding))
	}
	sort.Strings(parts)
	result.digest = baselineHash(strings.Join(parts, "\n") + "\n")
	return result, true
}

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
