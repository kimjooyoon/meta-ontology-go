package metricstrategyverify

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

const replayConceptBindingSchema = "gooo/concept-governed-strategy-binding/v1"

var replayConceptCarriers = map[string]string{
	"FOUNDATION": "bind-exact-source-metrics",
	"COHERENCE":  "project-algebraic-root-state",
	"REGRESSION": "replay-counterfactual",
}

var replayOperationConceptIDs = map[string]string{
	"bind-exact-source-metrics":       "metric-meta-program",
	"compact-obvious-lines":           "ci-selected-refactoring",
	"extract-function":                "ci-selected-refactoring",
	"exempt-project-root-topology":    "effect-bounded-observation",
	"interpret-dimension-registry":    "metric-meta-program",
	"lower-semantic-resolution":       "monotone-semantic-resolution",
	"observe-counterfactual-boundary": "effect-bounded-observation",
	"preserve-repository-workspace":   "effect-bounded-observation",
	"preserve-non-promoting-terminal": "rollback-fixed-point-recovery",
	"project-algebraic-root-state":    "metric-meta-program",
	"replay-counterfactual":           "causal-feedback-chain",
	"terminate-at-fixed-point":        "concept-governed-refactoring",
}

func replayConceptTrilemma(choice string) string {
	return map[string]string{"FOUNDATION": "AXIOM", "COHERENCE": "COHERENCE", "REGRESSION": "REGRESS"}[choice]
}

func replayConceptIndicatorID(indicator languageconcept.Indicator) string {
	suffix := strings.TrimPrefix(indicator.MetricID, "gooo.metric.meta.")
	return "gooo.strategy.concept." + strings.ToLower(indicator.Class) + "." + suffix
}

func replayConceptEvidenceDigest(fields map[string]string) (string, error) {
	fields["schema"] = replayConceptBindingSchema
	return artifact.Digest(fields)
}
