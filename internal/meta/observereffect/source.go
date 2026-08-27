package observereffect

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type semanticInterventionCase struct {
	Name     string
	Mutation string
	Suffix   string
}

var semanticInterventionCases = []semanticInterventionCase{
	{
		Name:     "comment-declaration-intervention",
		Mutation: "append comment-only entity and activity declarations",
		Suffix:   "\n// entity Fake id \"gooo://fake/entity\"\n// activity Fake(Entity) -> Entity\n",
	},
	{
		Name:     "quoted-text-comment-intervention",
		Mutation: "append quoted declaration-looking text inside a comment",
		Suffix:   "\n// \"entity Fake\" \"activity Fake(Entity) -> Entity\"\n",
	},
}

func canonicalSource(displayPath, filename string, payload []byte) Source {
	source := Source{Path: displayPath, Digest: DigestBytes(payload)}
	file, diagnostics := syntax.ParseFile(filename, string(payload))
	if file == nil || diagnostics.HasErrors() {
		source.Interventions = buildSemanticInterventions(filename, payload, "")
		return source
	}
	source.CanonicalParse = true
	ir, err := bidir.Lower(file)
	if err != nil {
		source.Interventions = buildSemanticInterventions(filename, payload, "")
		return source
	}
	source.CanonicalLowering = true
	source.SemanticDigest = ir.StableHash()
	source.GoooSource = true
	source.Interventions = buildSemanticInterventions(filename, payload, source.SemanticDigest)
	return source
}

func buildSemanticInterventions(filename string, payload []byte, baselineDigest string) []SemanticIntervention {
	interventions := make([]SemanticIntervention, 0, len(semanticInterventionCases))
	for _, intervention := range semanticInterventionCases {
		mutated := append(append([]byte(nil), payload...), []byte(intervention.Suffix)...)
		file, diagnostics := syntax.ParseFile(filename, string(mutated))
		parseValid := file != nil && !diagnostics.HasErrors()
		loweringValid := false
		mutatedDigest := ""
		if parseValid {
			ir, err := bidir.Lower(file)
			if err == nil {
				loweringValid = true
				mutatedDigest = ir.StableHash()
			}
		}
		invariant := baselineDigest != "" && parseValid && loweringValid && mutatedDigest == baselineDigest
		reason := "COMMENT_OR_QUOTED_TEXT_DID_NOT_CHANGE_SEMANTIC_IR"
		if !invariant {
			reason = "INTERVENTION_DID_NOT_REPLAY_CANONICAL_SEMANTICS"
		}
		interventions = append(interventions, SemanticIntervention{
			Name: intervention.Name, Mutation: intervention.Mutation,
			ParseValid: parseValid, LoweringValid: loweringValid,
			BaselineDigest: baselineDigest, MutatedDigest: mutatedDigest,
			SemanticInvariant: invariant,
			Producer:          "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "intervene-comment-and-quoted-text", ProofChoice: "REGRESSION",
			Stage: "BIND", Step: "parse-and-lower-intervention", Reason: reason,
		})
	}
	return interventions
}

func isCanonicalSource(source Source) bool {
	return strings.HasSuffix(source.Path, ".gooo") && source.GoooSource && source.CanonicalParse && source.CanonicalLowering && source.SemanticDigest != "" && len(source.Interventions) == len(semanticInterventionCases)
}
