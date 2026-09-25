package metricintervention

import (
	"fmt"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	verify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify"
)

func Generate(metricsPath, repository, subjectSHA string) (Ledger, error) {
	baseline, err := LoadBaseline(metricsPath, repository, subjectSHA)
	if err != nil {
		return Ledger{}, err
	}
	registry := DefaultRegistry()
	if err := ValidateRegistry(registry); err != nil {
		return Ledger{}, err
	}
	counterfactual, err := metric.Generate(repository, subjectSHA)
	if err != nil {
		return Ledger{}, err
	}
	replay, err := verify.Replay(counterfactual)
	if err != nil {
		return Ledger{}, err
	}
	predicted, err := Predict(counterfactual.Manifest, counterfactual.Plan)
	if err != nil {
		return Ledger{}, err
	}
	projections, err := Project(registry, baseline.Root, predicted, counterfactual.Delta)
	if err != nil {
		return Ledger{}, err
	}
	baselineDigest, err := artifact.Digest(baseline)
	if err != nil {
		return Ledger{}, err
	}
	registryDigest, err := artifact.Digest(registry)
	if err != nil {
		return Ledger{}, err
	}
	indicators, err := EvaluateIndicators(baseline, baselineDigest, registry, registryDigest, counterfactual, replay, projections)
	if err != nil {
		return Ledger{}, err
	}
	ledger := Ledger{Schema: LedgerSchema, Repository: repository, SubjectSHA: subjectSHA, ExecutionPolicy: "READ_ONLY_REPOSITORY_PLUS_DISPOSABLE_COUNTERFACTUAL", ScenarioKind: "ALGEBRAIC_ROOT_SCENARIO", Baseline: baseline, BaselineDigest: baselineDigest, Registry: registry, RegistryDigest: registryDigest, PredictedDelta: predicted, ObservedDelta: counterfactual.Delta, Counterfactual: counterfactual, CounterfactualVerification: replay, Projections: projections, Indicators: indicators, RepositoryWorkspaceWrites: false, PromotionAuthorized: false}
	if !AllSatisfied(indicators) {
		return Ledger{}, fmt.Errorf("metric intervention indicators are not satisfied")
	}
	return SealLedger(ledger)
}

func SealLedger(value Ledger) (Ledger, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

func ValidLedger(value Ledger) bool {
	digest := value.Digest
	sealed, err := SealLedger(value)
	return err == nil && digest == sealed.Digest
}

func AllSatisfied(indicators []Indicator) bool {
	for _, indicator := range indicators {
		if indicator.Status != "SATISFIED" {
			return false
		}
	}
	return len(indicators) > 0
}
