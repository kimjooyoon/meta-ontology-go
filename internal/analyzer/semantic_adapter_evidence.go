package analyzer

import (
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
