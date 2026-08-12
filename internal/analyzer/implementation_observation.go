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

func collectImplementationObservations(
	result Result, base semantic.IR, input SemanticAdapterInput,
) []ImplementationObservation {
	observations := make([]ImplementationObservation, 0)
	for _, fact := range result.Delta.Added {
		if fact.Origin != OriginImplementation {
			continue
		}
		observations = append(observations, ImplementationObservation{
			SourceDigest: input.SourceDigest, SourceFile: fact.Span.Filename,
			BaseDigest: base.StableHash(), PolicyDigest: input.Policy.Digest(),
			ToolchainDigest: input.ToolchainDigest, Subject: fact.Subject,
			Relation: fact.Relation, Object: fact.Object, Origin: fact.Origin,
			Span: fact.Span,
		})
	}
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Canonical() < observations[j].Canonical()
	})
	return observations
}

func implementationObservationDigest(observations []ImplementationObservation) string {
	var b strings.Builder
	b.WriteString(implementationObservationSchema)
	b.WriteString("/set\n")
	for _, observation := range observations {
		writeBindingField(&b, observation.Fingerprint())
	}
	return semantic.StableHashString(b.String())
}
