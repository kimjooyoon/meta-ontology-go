package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const protectedSlotSchema = "analyzer-protected-slot/v1"

const ProtectedSlotDeferred = "deferred-non-authoritative"

// ProtectedSlotObservation binds a generated handwritten slot to exact source
// bytes. It is deferred evidence and never becomes a semantic fact.
type ProtectedSlotObservation struct {
	SourceDigest    string `json:"source_digest"`
	SourceFile      string `json:"source_file"`
	BaseDigest      string `json:"base_digest"`
	PolicyDigest    string `json:"policy_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	RegistryDigest  string `json:"registry_digest"`
	SlotID          string `json:"slot_id"`
	Status          string `json:"status"`
	Span            Span   `json:"span"`
	BodySpan        Span   `json:"body_span"`
	BodyDigest      string `json:"body_digest"`
}

// Canonical returns the stable, source-bound slot record.
func (o ProtectedSlotObservation) Canonical() string {
	var builder strings.Builder
	builder.WriteString(protectedSlotSchema)
	builder.WriteByte('\n')
	writeBindingField(&builder, o.SourceDigest)
	writeBindingField(&builder, o.SourceFile)
	writeBindingField(&builder, o.BaseDigest)
	writeBindingField(&builder, o.PolicyDigest)
	writeBindingField(&builder, o.ToolchainDigest)
	writeBindingField(&builder, o.RegistryDigest)
	writeBindingField(&builder, o.SlotID)
	writeBindingField(&builder, o.Status)
	writeSlotSpan(&builder, o.Span)
	writeSlotSpan(&builder, o.BodySpan)
	writeBindingField(&builder, o.BodyDigest)
	return builder.String()
}

// Fingerprint returns the stable digest of this deferred slot observation.
func (o ProtectedSlotObservation) Fingerprint() string {
	return semantic.StableHashString(o.Canonical())
}

func protectedSlotObservationDigest(observations []ProtectedSlotObservation) string {
	var builder strings.Builder
	builder.WriteString(protectedSlotSchema)
	builder.WriteString("/set\n")
	ordered := append([]ProtectedSlotObservation(nil), observations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Canonical() < ordered[j].Canonical() })
	for _, observation := range ordered {
		writeBindingField(&builder, observation.Fingerprint())
	}
	return semantic.StableHashString(builder.String())
}

type protectedSlot struct {
	SourceFile string
	SlotID     string
	Span       Span
	BodySpan   Span
	BodyDigest string
}

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
