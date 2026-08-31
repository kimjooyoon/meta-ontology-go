package verticalsliceclosureshadow

import (
	"encoding/json"

	languagesemantic "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic"
	languagesyntax "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func syntaxFixture(head string) []byte {
	return fixtureJSON(map[string]any{"schema": "gooo/language-syntax-roundtrip/v1",
		"decision": "PASS", "resolution": "EXACT", "head_sha": head,
		"meta_operation": "prove-language-syntax-roundtrip",
		"report_digest":  fixtureDigest("a"), "repository_writes": 0,
		"mutation_authorized": false,
		"source": map[string]any{"expected_head_sha": head,
			"concept_artifact_digest": fixtureDigest("1")},
		"summary": map[string]any{"satisfied": languagesyntax.FixedTotal, "total": languagesyntax.FixedTotal, "executed": languagesyntax.FixedTotal,
			"capability_satisfied": languagesyntax.FixedCapabilityTotal, "capability_total": languagesyntax.FixedCapabilityTotal, "capability_executed": languagesyntax.FixedCapabilityTotal, "capability_unresolved": 0,
			"governance_satisfied": languagesyntax.FixedGovernanceTotal, "governance_total": languagesyntax.FixedGovernanceTotal, "governance_executed": languagesyntax.FixedGovernanceTotal, "governance_unresolved": 0,
			"not_satisfied": 0, "unresolved": 0, "readiness_bps": 10000}})
}

func semanticFixture(head string, syntax []byte) []byte {
	return fixtureJSON(map[string]any{"schema": "gooo/language-semantic-model/v1",
		"decision": "PASS", "resolution": "EXACT", "report_digest": fixtureDigest("b"),
		"repository_writes": 0, "mutation_authorized": false,
		"source": map[string]any{"expected_head_sha": head,
			"meta_operation":         "prove-staged-semantic-model",
			"syntax_artifact_digest": digestBytes(syntax),
			"syntax_report_digest":   fixtureDigest("a")},
		"summary": map[string]any{"satisfied": languagesemantic.FixedTotal, "total": languagesemantic.FixedTotal, "executed": languagesemantic.FixedTotal,
			"not_satisfied": 0, "unresolved": 0, "readiness_bps": 10000,
			"stage_order_violations": 0, "effectful_stages": 0, "registry_drift": 0}})
}

func bindingFixture(head string, semantic []byte) []byte {
	return fixtureJSON(map[string]any{
		"schema":   "gooo/language-semantic-readiness-binding/v2",
		"decision": "PASS", "resolution": "EXACT", "report_digest": fixtureDigest("c"),
		"repository_writes": 0, "mutation_authorized": false,
		"source": map[string]any{"expected_head_sha": head,
			"meta_operation":         "bind-semantic-readiness-evidence",
			"semantic_file_digest":   digestBytes(semantic),
			"semantic_report_digest": fixtureDigest("b")},
		"summary": map[string]any{"coordinates": 12, "bound_coordinates": 12,
			"unresolved": 0, "readiness_completed": 21, "readiness_total": 24,
			"readiness_bps": 8750, "semantic_satisfied": languagesemantic.FixedTotal, "semantic_total": languagesemantic.FixedTotal,
			"effectful_stages": 0, "mutation_authorities": 0}})
}

func useCaseFixture(head string, syntax []byte) []byte {
	var parsed artifactEnvelope
	_ = json.Unmarshal(syntax, &parsed)
	return fixtureJSON(map[string]any{"schema": "gooo/toolchain-executable-use-cases/v1",
		"decision": "PASS", "resolution": "EXACT", "head_sha": head,
		"meta_operation": "execute-versioned-use-cases", "report_digest": fixtureDigest("d"),
		"repository_writes": 0, "mutation_authorized": false,
		"source": map[string]any{"expected_head_sha": head,
			"concept_artifact_digest": parsed.Source.ConceptArtifactDigest},
		"summary": map[string]any{"satisfied": 3, "total": 3, "executed": 3,
			"not_satisfied": 0, "unresolved": 0, "readiness_bps": 10000}})
}

func fixtureJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}
