package analyzer

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const implementationObservationSchema = "analyzer-implementation-observation/v1"

// ImplementationObservation is a deferred source observation. It is not a
// semantic fact and cannot alter IR authority without an explicit mapping.
type ImplementationObservation struct {
	SourceDigest    string
	SourceFile      string
	BaseDigest      string
	PolicyDigest    string
	ToolchainDigest string
	RegistryDigest  string
	Subject         Identity
	Relation        Relation
	Object          Identity
	Origin          ObservationOrigin
	Span            Span
}

// Canonical binds the observation to exact source, file identity, semantic
// base, policy, toolchain, stable IDs, relation, origin, and source span.
func (o ImplementationObservation) Canonical() string {
	var b strings.Builder
	b.WriteString(implementationObservationSchema)
	b.WriteByte('\n')
	writeBindingField(&b, o.SourceDigest)
	writeBindingField(&b, o.SourceFile)
	writeBindingField(&b, o.BaseDigest)
	writeBindingField(&b, o.PolicyDigest)
	writeBindingField(&b, o.ToolchainDigest)
	writeBindingField(&b, o.RegistryDigest)
	writeBindingField(&b, o.Subject.Namespace)
	writeBindingField(&b, o.Subject.ID)
	writeBindingField(&b, string(o.Relation))
	writeBindingField(&b, o.Object.Namespace)
	writeBindingField(&b, o.Object.ID)
	writeBindingField(&b, string(o.Origin))
	writeBindingField(&b, o.Span.Filename)
	writeBindingField(&b, intString(o.Span.Start.Offset))
	writeBindingField(&b, intString(o.Span.Start.Line))
	writeBindingField(&b, intString(o.Span.Start.Column))
	writeBindingField(&b, intString(o.Span.End.Offset))
	writeBindingField(&b, intString(o.Span.End.Line))
	writeBindingField(&b, intString(o.Span.End.Column))
	return b.String()
}

// Fingerprint returns the stable digest of this non-authoritative observation.
func (o ImplementationObservation) Fingerprint() string {
	return semantic.StableHashString(o.Canonical())
}

func validImplementationObservation(observation ImplementationObservation) bool {
	if !validDigest(observation.SourceDigest) || !validDigest(observation.BaseDigest) ||
		!validDigest(observation.PolicyDigest) || !validDigest(observation.ToolchainDigest) ||
		!validDigest(observation.RegistryDigest) || observation.SourceFile == "" ||
		observation.Span.Filename != observation.SourceFile ||
		observation.Span.Start.Offset < 0 ||
		observation.Span.End.Offset < observation.Span.Start.Offset ||
		observation.Origin != OriginImplementation || !knownAnalyzerRelation(observation.Relation) ||
		!observation.Subject.Valid() || !observation.Object.Valid() {
		return false
	}
	if _, err := semantic.ParseIdentity(observation.Subject.ID); err != nil {
		return false
	}
	if _, err := semantic.ParseIdentity(observation.Object.ID); err != nil {
		return false
	}
	return true
}

func collectImplementationObservations(
	result Result, base semantic.IR, input SemanticAdapterInput,
) []ImplementationObservation {
	observations := make([]ImplementationObservation, 0)
	for _, fact := range result.Delta.Added {
		if fact.Origin != OriginImplementation {
			continue
		}
		mapping, mapped := input.Policy.lookup(fact.Relation)
		if mapped && mapping.allowsOrigin(fact.Origin) {
			continue
		}
		observations = append(observations, ImplementationObservation{
			SourceDigest: input.SourceDigest, SourceFile: fact.Span.Filename,
			BaseDigest: base.StableHash(), PolicyDigest: input.Policy.Digest(),
			ToolchainDigest: input.ToolchainDigest, RegistryDigest: input.Registry.Digest(), Subject: fact.Subject,
			Relation: fact.Relation, Object: fact.Object, Origin: fact.Origin,
			Span: fact.Span,
		})
	}
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Canonical() < observations[j].Canonical()
	})
	return observations
}

func implementationObservationsMatch(
	left, right []ImplementationObservation,
) bool {
	orderedLeft := append([]ImplementationObservation(nil), left...)
	orderedRight := append([]ImplementationObservation(nil), right...)
	sort.Slice(orderedLeft, func(i, j int) bool {
		return orderedLeft[i].Canonical() < orderedLeft[j].Canonical()
	})
	sort.Slice(orderedRight, func(i, j int) bool {
		return orderedRight[i].Canonical() < orderedRight[j].Canonical()
	})
	if len(orderedLeft) != len(orderedRight) {
		return false
	}
	for index := range orderedLeft {
		if orderedLeft[index].Canonical() != orderedRight[index].Canonical() {
			return false
		}
	}
	return true
}

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
