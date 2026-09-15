package coupling

import (
	"sort"
	"strings"
)

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
