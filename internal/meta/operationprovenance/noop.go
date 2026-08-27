package operationprovenance

import "fmt"

func RejectNoopMutation(source []byte, artifactRoot string) (Issue, error) {
	ir, err := lowerSource(source)
	if err != nil {
		return Issue{}, err
	}
	metrics, scenarios, _, _, err := reconstructSemanticData(ir)
	if err != nil {
		return Issue{}, err
	}
	if _, err := validateContract(metrics, scenarios); err != nil {
		return Issue{}, err
	}
	artifacts, err := collectArtifacts(artifactRoot, metrics)
	if err != nil {
		return Issue{}, err
	}
	noop := scenarioSpec{ID: "noop-probe", RemoveRelation: "CONSUMES:DOES-NOT-EXIST", Reason: "probe must be rejected"}
	if _, err := evaluateScenario(metrics, noop, artifacts, digestBytes(source), "probe"); err == nil {
		return Issue{}, fmt.Errorf("no-op mutation was accepted")
	} else {
		return Issue{Stage: "SCENARIO", Step: "apply-mutation", Reason: "NOOP_MUTATION_REJECTED", Detail: err.Error(), Cause: "DIRECT_CAUSE"}, nil
	}
}
