package operationprovenance

func evaluateScenario(metrics []metricSpec, scenario scenarioSpec, observations map[string][]RelationObservation, sourceDigest, semanticDigest string) (ScenarioResult, error) {
	working, err := fixtureFromArtifacts(metrics, scenario, observations)
	if err != nil {
		return ScenarioResult{}, err
	}
	return evaluateFixture(working, sourceDigest, semanticDigest), nil
}
