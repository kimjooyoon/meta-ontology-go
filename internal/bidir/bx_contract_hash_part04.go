package bidir

import (
	"fmt"
	"sort"
	"strings"
)

func baseEvidence(input BXBaseEvidenceInput, document Document, model Model) (BXBaseEvidence, error) {
	if documentDigest(input.DSL) != documentDigest(document) {
		return BXBaseEvidence{}, fmt.Errorf("base DSL artifact does not match fixture document")
	}
	if !SemanticEquivalent(input.IR, model) {
		return BXBaseEvidence{}, fmt.Errorf("base IR artifact does not match fixture model")
	}
	if len(input.Go) == 0 || len(input.SourceMap) == 0 || len(input.Evidence) == 0 || len(input.Provenance) == 0 {
		return BXBaseEvidence{}, fmt.Errorf("base evidence requires non-empty Go, source-map, evidence, and provenance artifacts")
	}
	return BXBaseEvidence{
		DSL:        artifact(documentDigest(input.DSL), 1),
		IR:         artifact(SemanticFingerprint(input.IR), len(input.IR.Nodes)+len(input.IR.Relations)),
		Go:         artifact(digestFacts(input.Go), len(input.Go)),
		SourceMap:  artifact(digestSpans(input.SourceMap), len(input.SourceMap)),
		Evidence:   artifact(digestFacts(input.Evidence), len(input.Evidence)),
		Provenance: artifact(digestSpans(input.Provenance), len(input.Provenance)),
	}, nil
}
func digestFacts(facts FactSet) string {
	copySet := facts.Normalized()
	var builder strings.Builder
	for _, fact := range copySet {
		writePart(&builder, factCanonical(fact))
	}
	return digest(builder.String())
}
func digestSpans(spans []SourceSpan) string {
	copySpans := append([]SourceSpan(nil), spans...)
	sort.Slice(copySpans, func(i, j int) bool { return spanLess(copySpans[i], copySpans[j]) })
	var builder strings.Builder
	for _, span := range copySpans {
		writeSpan(&builder, span)
	}
	return digest(builder.String())
}
