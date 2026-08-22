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
		concept := conceptEvidence{}
		concept.ID = id
		concept.Stage = "OPERATING"
		concept.CodeBindings = []string{"internal/meta/example"}
		concept.MetricBindings = []string{"gooo.metric.example.v1"}
		concept.UseCases = []useCaseEvidence{{ID: "explicit-case", Trigger: "explicit input", ExpectedOutcome: "PASS"}}
		concepts = append(concepts, concept)
	}
	artifact := conceptArtifact{}
	artifact.Schema = conceptArtifactSchema
	artifact.Decision = decision
	artifact.CatalogDigest = "sha256:catalog"
	artifact.ReplayReportDigest = "sha256:report"
	artifact.ReplayEqual = true
	artifact.Report = conceptReport{Concepts: concepts, ReportDigest: "sha256:report"}
	artifact.ArtifactDigest = "sha256:artifact"
	raw, err := json.Marshal(artifact)
	if err != nil {
		panic(err)
	}
	return raw
}
