package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func implementationObservationDigest(
	observations []ImplementationObservation, slots []ProtectedSlotObservation,
) string {
	var b strings.Builder
	b.WriteString(implementationObservationSchema)
	b.WriteString("/set\n")
	orderedObservations := append([]ImplementationObservation(nil), observations...)
	sort.Slice(orderedObservations, func(i, j int) bool {
		return orderedObservations[i].Canonical() < orderedObservations[j].Canonical()
	})
	for _, observation := range orderedObservations {
		writeBindingField(&b, "implementation")
		writeBindingField(&b, observation.Fingerprint())
	}
	orderedSlots := append([]ProtectedSlotObservation(nil), slots...)
	sort.Slice(orderedSlots, func(i, j int) bool {
		return orderedSlots[i].Canonical() < orderedSlots[j].Canonical()
	})
	for _, slot := range orderedSlots {
		writeBindingField(&b, "protected-slot")
		writeBindingField(&b, slot.Fingerprint())
	}
	return semantic.StableHashString(b.String())
}
