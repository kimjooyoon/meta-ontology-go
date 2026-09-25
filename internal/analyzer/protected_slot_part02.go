package analyzer

import (
	"fmt"
	"sort"
	"strings"
)

type protectedSlotMarker struct {
	kind    string
	id      string
	start   int
	lineEnd int
	next    int
}

func collectProtectedSlots(sources []SourceFile) ([]protectedSlot, error) {
	ordered, err := canonicalSourceFiles(sources)
	if err != nil {
		return nil, err
	}
	var slots []protectedSlot
	seen := make(map[string]string)
	for _, source := range ordered {
		fileSlots, err := parseProtectedSlots(source)
		if err != nil {
			return nil, err
		}
		for _, slot := range fileSlots {
			if previous, exists := seen[slot.SlotID]; exists {
				return nil, adapterError(AdapterSlotConfig, "", slot.SlotID,
					fmt.Sprintf("duplicate slot identity; first declared in %s", previous))
			}
			seen[slot.SlotID] = source.Filename
			slots = append(slots, slot)
		}
	}
	sort.Slice(slots, func(i, j int) bool {
		if slots[i].SourceFile != slots[j].SourceFile {
			return slots[i].SourceFile < slots[j].SourceFile
		}
		return slots[i].Span.Start.Offset < slots[j].Span.Start.Offset
	})
	return slots, nil
}
func bindProtectedSlots(
	slots []protectedSlot, sourceDigest, baseDigest, policyDigest, toolchainDigest,
	registryDigest string,
) []ProtectedSlotObservation {
	observations := make([]ProtectedSlotObservation, 0, len(slots))
	for _, slot := range slots {
		observations = append(observations, ProtectedSlotObservation{
			SourceDigest: sourceDigest, SourceFile: slot.SourceFile,
			BaseDigest: baseDigest, PolicyDigest: policyDigest,
			ToolchainDigest: toolchainDigest, RegistryDigest: registryDigest, SlotID: slot.SlotID,
			Status: ProtectedSlotDeferred, Span: slot.Span,
			BodySpan: slot.BodySpan, BodyDigest: slot.BodyDigest,
		})
	}
	sort.Slice(observations, func(i, j int) bool { return observations[i].Canonical() < observations[j].Canonical() })
	return observations
}
func writeSlotSpan(builder *strings.Builder, span Span) {
	writeBindingField(builder, span.Filename)
	writeBindingField(builder, intString(span.Start.Offset))
	writeBindingField(builder, intString(span.Start.Line))
	writeBindingField(builder, intString(span.Start.Column))
	writeBindingField(builder, intString(span.End.Offset))
	writeBindingField(builder, intString(span.End.Line))
	writeBindingField(builder, intString(span.End.Column))
}
