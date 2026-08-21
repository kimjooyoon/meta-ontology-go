package coupling

import (
	"sort"
)

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
