package proposal

import artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"

func buildIndicators(summary Summary, actions, writes int, promotion bool) []Indicator {
	producer, consumer := "proposal.Evaluate", "language-readiness"
	promotionValue := 0
	if promotion { promotionValue = 1 }
	return []Indicator{
		{MetricID: "gooo.metric.language.change-proposal-contract-bps.v1", Class: "outcome", ProofChoice: "COHERENCE", Producer: producer, Consumer: consumer, MetaOperation: "quantify-change-proposal-contract", Value: summary.ReadinessBPS, Target: 10000, Unit: "BASIS_POINT", Satisfied: summary.ReadinessBPS == 10000},
		{MetricID: "gooo.metric.language.change-proposal-satisfied-coordinates.v1", Class: "driver", ProofChoice: "FOUNDATION", Producer: producer, Consumer: consumer, MetaOperation: "count-satisfied-proposal-coordinates", Value: summary.Satisfied, Target: len(registry), Unit: "COORDINATE", Satisfied: summary.Satisfied == len(registry)},
		{MetricID: "gooo.metric.language.change-proposal-selected-actions.v1", Class: "driver", ProofChoice: "COHERENCE", Producer: producer, Consumer: consumer, MetaOperation: "count-independent-proposal-actions", Value: actions, Target: 2, Unit: "ACTION", Satisfied: actions == 2},
		{MetricID: "gooo.metric.language.change-proposal-unresolved.guardrail.v1", Class: "guardrail", ProofChoice: "REGRESSION", Producer: producer, Consumer: consumer, MetaOperation: "lower-resolution-on-unknown-proposal", Value: summary.Unresolved, Target: 0, Unit: "COORDINATE", Satisfied: summary.Unresolved == 0},
		{MetricID: "gooo.metric.language.change-proposal-repository-writes.guardrail.v1", Class: "guardrail", ProofChoice: "FOUNDATION", Producer: producer, Consumer: consumer, MetaOperation: "preserve-read-only-proposal", Value: writes, Target: 0, Unit: "REPOSITORY_WRITE", Satisfied: writes == 0},
		{MetricID: "gooo.metric.language.change-proposal-promotion-authorized.guardrail.v1", Class: "guardrail", ProofChoice: "FOUNDATION", Producer: producer, Consumer: consumer, MetaOperation: "deny-proposal-promotion-authority", Value: promotionValue, Target: 0, Unit: "AUTHORIZATION", Satisfied: promotionValue == 0},
	}
}

func buildProofs(coordinates []Coordinate) ([]Proof, error) {
	choices := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	result := make([]Proof, 0, len(choices))
	for _, choice := range choices {
		passed := true
		evidence := make([]string, 0)
		for _, coordinate := range coordinates {
			if coordinate.ProofChoice == choice { passed = passed && coordinate.Status == "SATISFIED"; evidence = append(evidence, coordinate.EvidenceDigest) }
		}
	digest, err := artifact.Digest(evidence)
		if err != nil { return nil, err }
		result = append(result, Proof{Choice: choice, MetaOperation: "justify-change-proposal-by-" + choice, Passed: passed, EvidenceDigest: digest})
	}
	return result, nil
}
