package analyzer

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const shadowedCandidateEvidenceSchema = "analyzer-shadowed-candidate-evidence/v1"

// ShadowedCandidateEvidenceCanonical returns a deterministic, non-authoritative
// evidence stream for candidate observations that were shadowed by an existing
// deterministic fact. It is deliberately separate from semantic.IR.Canonical:
// these observations must remain review evidence and must not change authority.
func (result SemanticAdapterResult) ShadowedCandidateEvidenceCanonical() string {
	records := append([]semantic.Evidence(nil), result.ShadowedCandidateEvidence...)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Canonical() < records[j].Canonical()
	})
	var b strings.Builder
	b.WriteString(shadowedCandidateEvidenceSchema)
	b.WriteByte('\n')
	for _, record := range records {
		b.WriteString(record.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}

// ShadowedCandidateEvidenceHash is the stable digest of the non-authoritative
// shadow stream. An empty stream still has a defined schema-bound digest.
func (result SemanticAdapterResult) ShadowedCandidateEvidenceHash() string {
	return semantic.StableHashString(result.ShadowedCandidateEvidenceCanonical())
}

func mappedEvidence(input SemanticAdapterInput, source Relation, fact semantic.Fact, status semantic.FactStatus) (semantic.Evidence, error) {
	identitySeed := strings.Join([]string{
		input.Policy.Revision,
		string(source),
		fact.Key().Subject.String(),
		fact.Key().Predicate.String(),
		fact.Key().Object.String(),
		status.String(),
		fact.Span.File,
		intString(fact.Span.Start.Offset),
		intString(fact.Span.End.Offset),
	}, "\x00")
	evidenceID := semantic.ID("gooo://evidence/analyzer/" + semantic.StableHashString(identitySeed))
	evidence, err := semantic.NewEvidence(
		evidenceID,
		input.Producer,
		input.EvidenceKind,
		fact.Key(),
		input.SourceDigest,
	)
	if err != nil {
		return semantic.Evidence{}, err
	}
	evidence.Status = status
	return evidence.WithSpan(fact.Span), nil
}

func intString(value int) string {
	return strconv.Itoa(value)
}
