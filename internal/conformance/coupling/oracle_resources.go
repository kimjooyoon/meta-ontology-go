package coupling

func normalizeResources(receipts []ExternalResourceReceipt, expected ResourceBindingConfig) (ResourceObservation, oracleValidation) {
	seen := make(map[string]struct{}, len(receipts))
	var out ResourceObservation
	for _, receipt := range receipts {
		if !validID(receipt.ReceiptID) || !validID(expected.ProviderID) || !validID(expected.ObserverID) || !validDigest(expected.ProviderDigest) || !validDigest(expected.ObserverDigest) || !validDigest(expected.SnapshotDigest) || !validDigest(expected.SourceDigest) || !validDigest(receipt.ProviderDigest) || !validDigest(receipt.ObserverDigest) || !validDigest(receipt.SnapshotDigest) || !validDigest(receipt.SourceDigest) || !validDigest(receipt.BindingDigest) ||
			expected.ProviderDigest != resourceProviderDigest(expected.ProviderID) || expected.ObserverDigest != resourceObserverDigest(expected.ObserverID) || expected.SourceDigest != resourceSourceDigest(expected.ProviderID, expected.ObserverID, expected.SnapshotDigest) ||
			receipt.ProviderDigest != expected.ProviderDigest || receipt.ObserverDigest != expected.ObserverDigest || receipt.SnapshotDigest != expected.SnapshotDigest || receipt.SourceDigest != expected.SourceDigest ||
			!receipt.Present || !receipt.Independent || receipt.State != "CURRENT" ||
			receipt.BindingDigest != resourceBindingDigest(receipt) {
			return ResourceObservation{}, oracleValidation{DecisionUnknown, ReasonResourceUnbound}
		}
		if _, duplicate := seen[receipt.Metric]; duplicate {
			return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
		}
		seen[receipt.Metric] = struct{}{}
		switch receipt.Metric {
		case "cpu-core-ns":
			if receipt.Unit != "ns" {
				return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
			}
			out.CPUCoreNS = receipt.Value
		case "peak-memory-bytes":
			if receipt.Unit != "bytes" {
				return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
			}
			out.PeakMemoryBytes = receipt.Value
		case "work-units":
			if receipt.Unit != "units" {
				return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
			}
			out.WorkUnits = receipt.Value
		default:
			return ResourceObservation{}, oracleValidation{DecisionFailClosed, ReasonResourceUnbound}
		}
	}
	if len(seen) != 3 {
		return ResourceObservation{}, oracleValidation{DecisionUnknown, ReasonResourceUnbound}
	}
	return out, oracleValidation{}
}

func resourceBindingsEqual(left, right ResourceBindingConfig) bool {
	return left.ProviderID != "" && left.ProviderID == right.ProviderID && left.ObserverID != "" && left.ObserverID == right.ObserverID && left.ProviderDigest != "" && left.ProviderDigest == right.ProviderDigest && left.ObserverDigest != "" && left.ObserverDigest == right.ObserverDigest && left.SnapshotDigest != "" && left.SnapshotDigest == right.SnapshotDigest && left.SourceDigest != "" && left.SourceDigest == right.SourceDigest
}

func validateManifest(input Input, beforeDigest, afterDigest, registryDigest string) oracleValidation {
	manifest := input.Manifest
	if !manifest.Complete {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	expectedBefore := stateSnapshotDigest(input.AuthoritySourceBefore, beforeDigest, registryDigest, input.Config)
	expectedAfter := stateSnapshotDigest(input.AuthoritySourceAfter, afterDigest, registryDigest, input.Config)
	if manifest.BeforeSnapshotDigest != expectedBefore || manifest.AfterSnapshotDigest != expectedAfter || manifest.ToolchainDigest != input.Config.ToolchainDigest || manifest.ProfileDigest != input.Config.Profile.Digest || manifest.RegistryDigest != registryDigest {
		return oracleValidation{DecisionFailClosed, ReasonDigestMismatch}
	}
	if manifest.ZeroChange {
		if len(input.Changes) != 0 || beforeDigest != afterDigest || manifest.BeforeSnapshotDigest != manifest.AfterSnapshotDigest || len(input.Receipts) != 0 || len(input.Path.Edges) != 0 || len(input.Path.Claims) != 0 || len(input.Path.Evidence) != 0 || len(input.Roots) != 0 {
			return oracleValidation{DecisionFailClosed, ReasonChangedSurface}
		}
	}
	return oracleValidation{}
}
