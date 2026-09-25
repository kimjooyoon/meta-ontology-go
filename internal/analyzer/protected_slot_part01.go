package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
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
