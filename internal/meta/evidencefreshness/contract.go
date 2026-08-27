package evidencefreshness

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"

func CanonicalContract() model.Contract {
	return model.Contract{
		Schema:     model.ContractSchema,
		Scope:      model.Scope,
		SourcePath: "examples/evidence-freshness/main.gooo",
		BaseContext: model.Context{
			Schema: model.ContextSchema,
			Tuple: model.EvidenceTuple{
				Recipe:      "recipe:gooo-evidence-freshness/v2",
				Environment: "environment:go1.27/linux-amd64/hermetic",
				Runner:      "runner:github-actions/ubuntu-24.04",
				Verifier:    "verifier:evidence-freshness-decider/v2",
			},
			CurrentEpoch:        20260827,
			EnvironmentBoundary: "environment:go1.27/linux-amd64/hermetic",
			Consumer:            model.ConsumerID,
		},
		Metrics: []model.MetricDefinition{
			metric("gooo.metric.evidence-freshness.cases.v2", "OUTCOME", "FOUNDATION", "observe-freshness-cases", model.CaseTotal, model.CaseTotal),
			metric("gooo.metric.evidence-freshness.current-evidence.v2", "OUTCOME", "FOUNDATION", "observe-current-evidence", model.CurrentEvidenceTotal, model.CurrentEvidenceTotal),
			metric("gooo.metric.evidence-freshness.synthetic-counterexamples.v2", "DRIVER", "COHERENCE", "observe-synthetic-counterexamples", model.SyntheticCounterexampleTotal, model.SyntheticCounterexampleTotal),
			metric("gooo.metric.evidence-freshness.coupling-axes.v2", "DRIVER", "COHERENCE", "enumerate-six-coupling-axes", model.AxisTotal, model.AxisTotal),
			metric("gooo.metric.evidence-freshness.raw-source-reconstruction.v2", "OUTCOME", "FOUNDATION", "reconstruct-raw-source", 9, 9),
			metric("gooo.metric.evidence-freshness.semantic-source-reconstruction.v2", "OUTCOME", "FOUNDATION", "reconstruct-canonical-semantic-source", 9, 9),
			metric("gooo.metric.evidence-freshness.comment-intervention.v2", "GUARDRAIL", "COHERENCE", "preserve-comment-insensitive-semantic-claim", 1, 1),
			metric("gooo.metric.evidence-freshness.semantic-intervention.v2", "GUARDRAIL", "REGRESSION", "detect-semantic-value-change", 1, 1),
			metric("gooo.metric.evidence-freshness.freshness-transitions.v2", "OUTCOME", "FOUNDATION", "record-freshness-observation-transitions", model.TransitionTotal, model.TransitionTotal),
			metric("gooo.metric.evidence-freshness.claim-ledger.v2", "OUTCOME", "FOUNDATION", "append-claim-ledger-chain", model.TransitionTotal, model.TransitionTotal),
			metric("gooo.metric.evidence-freshness.source-unavailable.v2", "GUARDRAIL", "REGRESSION", "lower-source-unavailable-evidence", 1, 1),
			metric("gooo.metric.evidence-freshness.read-only-before-after.v2", "GUARDRAIL", "REGRESSION", "bind-ci-before-after-write-set", 1, 1),
			metric("gooo.metric.evidence-freshness.independence-contract.v2", "GUARDRAIL", "FOUNDATION", "verify-producer-import-boundary", model.IndependenceContractTotal, model.IndependenceContractTotal),
		},
		NotClaimed: []string{
			"cryptographic signature authenticity",
			"full compiler semantic correctness",
			"general cache eviction or recomputation",
			"wall-clock time or scheduler freshness",
			"external side effects or mutation authority",
		},
	}
}

func metric(id, class, proof, operation string, expected, denominator int) model.MetricDefinition {
	return model.MetricDefinition{MetricID: id, Class: class, Producer: model.ProducerID,
		Consumer: model.ConsumerID, ProofChoice: proof, MetaOperation: operation,
		ExpectedNumerator: expected, Denominator: denominator}
}
