package proposal

import (
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

func Evaluate(repository, subject string, first, replay strategy.Plan, receipt strategyverify.Receipt) (Report, error) {
	strategyCoordinates, strategyFacts, err := strategyCoordinates(repository, subject, first, replay, receipt)
	if err != nil {
		return Report{}, err
	}
	generationCoordinates, generationFacts, err := generationCoordinates()
	if err != nil {
		return Report{}, err
	}
	writes := boolCount(strategyFacts.Writes, generationFacts.Writes)
	promotion := strategyFacts.Promotion || generationFacts.Promotion
	boundaryOK := writes == 0 && !promotion
	status, reason := coordinateStatus(boundaryOK, false, "READ_ONLY_NON_AUTHORIZING_BOUNDARY_PROVEN")
	boundary, err := makeCoordinate(7, status, reason, []any{writes, promotion, strategyFacts.PlanDigest, strategyFacts.ReceiptDigest})
	if err != nil {
		return Report{}, err
	}
	coordinates := append(strategyCoordinates, generationCoordinates...)
	coordinates = append(coordinates, boundary)
	summary := summarize(coordinates)
	decision, decisionReason := decisionFor(summary)
	digest, err := registryDigest()
	if err != nil {
		return Report{}, err
	}
	proofs, err := buildProofs(coordinates)
	if err != nil {
		return Report{}, err
	}
	report := Report{Schema: Schema, RegistrySchema: RegistrySchema, RegistryDigest: digest, Repository: repository, SubjectSHA: subject, Decision: decision, Reason: decisionReason, StrategyDecision: strategyFacts.Decision, ProposalDecision: generationFacts.Decision, SelectedActions: generationFacts.Actions, Summary: summary, Coordinates: coordinates, Indicators: buildIndicators(summary, generationFacts.Actions, writes, promotion), Proofs: proofs, RepositoryWrites: writes, PromotionAuthorized: promotion}
	report, err = sealReport(report)
	if err != nil {
		return Report{}, err
	}
	return report, Validate(report)
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}
