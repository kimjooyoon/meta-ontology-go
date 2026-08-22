package languagereadiness

import "encoding/json"

var currentConceptIDs = []string{
	"metric-meta-program",
	"executable-actionability",
	"effect-bounded-observation",
	"monotone-semantic-resolution",
	"causal-feedback-chain",
	"ci-selected-refactoring",
	"concept-governed-refactoring",
}

func artifactFixture(decision string, conceptIDs ...string) []byte {
	concepts := make([]conceptEvidence, 0, len(conceptIDs))
	for _, id := range conceptIDs {
		concepts = append(concepts, conceptEvidence{
			ID: id, Stage: "OPERATING",
			CodeBindings: []string{"internal/meta/example"},
			MetricBindings: []string{"gooo.metric.example.v1"},
			UseCases: []useCaseEvidence{{
				ID: "explicit-case", Trigger: "explicit input", ExpectedOutcome: "PASS",
			}},
		})
	}
	artifact := conceptArtifact{
		Schema: conceptArtifactSchema, Decision: decision,
		CatalogDigest: "sha256:catalog", ReplayReportDigest: "sha256:report", ReplayEqual: true,
		Report: conceptReport{Concepts: concepts, ReportDigest: "sha256:report"},
		ArtifactDigest: "sha256:artifact",
	}
	raw, err := json.Marshal(artifact)
	if err != nil {
		panic(err)
	}
	return raw
}
