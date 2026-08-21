package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validProtectedSlotObservation(observation ProtectedSlotObservation) bool {
	if !validDigest(observation.SourceDigest) || !validDigest(observation.BaseDigest) ||
		!validDigest(observation.PolicyDigest) || !validDigest(observation.ToolchainDigest) ||
		!validDigest(observation.RegistryDigest) || !validDigest(observation.BodyDigest) ||
		observation.SourceFile == "" || observation.SlotID == "" ||
		observation.Status != ProtectedSlotDeferred {
		return false
	}
	if _, err := semantic.ParseIdentity(observation.SlotID); err != nil {
		return false
	}
	if observation.Span.Filename != observation.SourceFile || observation.BodySpan.Filename != observation.SourceFile {
		return false
	}
	if !validProtectedSpan(observation.Span) || !validProtectedSpan(observation.BodySpan) {
		return false
	}
	return observation.Span.Start.Offset <= observation.BodySpan.Start.Offset &&
		observation.BodySpan.End.Offset <= observation.Span.End.Offset
}
func validProtectedSpan(span Span) bool {
	return span.Filename != "" && span.Start.Offset >= 0 &&
		span.End.Offset >= span.Start.Offset
}
func validateSlotObservations(
	slots []ProtectedSlotObservation, sourceDigest, baseDigest, policyDigest, toolchainDigest,
	registryDigest string,
) error {
	seen := make(map[string]struct{}, len(slots))
	for _, slot := range slots {
		if !validProtectedSlotObservation(slot) {
			return adapterError(AdapterSlotConfig, "", slot.SlotID, "protected slot observation is invalid")
		}
		if slot.SourceDigest != sourceDigest || slot.BaseDigest != baseDigest ||
			slot.PolicyDigest != policyDigest || slot.ToolchainDigest != toolchainDigest ||
			slot.RegistryDigest != registryDigest {
			return adapterError(AdapterSlotConfig, "", slot.SlotID, "protected slot binding does not match adapter input")
		}
		if _, exists := seen[slot.SlotID]; exists {
			return adapterError(AdapterSlotConfig, "", slot.SlotID, "duplicate protected slot observation")
		}
		seen[slot.SlotID] = struct{}{}
	}
	return nil
}
