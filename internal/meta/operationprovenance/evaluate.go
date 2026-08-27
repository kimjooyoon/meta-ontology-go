package operationprovenance

func evaluateFixture(f fixture, sourceDigest, semanticDigest string) ScenarioResult {
	byID := make(map[string]metricSpec, len(f.Metrics))
	for _, metric := range f.Metrics {
		byID[metric.ID] = metric
	}
	evaluator := evaluator{fixture: f, byID: byID, memo: map[string]MetricResult{}, visiting: map[string]bool{}, sourceDigest: sourceDigest, semanticDigest: semanticDigest}
	results := make([]MetricResult, 0, len(f.Metrics))
	decisions, transitions := map[string]int{}, map[string]int{}
	numerator := 0
	lineageExact := true
	for _, metric := range f.Metrics {
		result := evaluator.run(metric.ID)
		results = append(results, result)
		decisions[result.Decision]++
		transitions[result.Transition.Transition]++
		numerator += result.Numerator
		lineageExact = lineageExact && result.LineageResolution == "EXACT"
	}
	return ScenarioResult{ID: f.ID, Mutation: f.Mutation, Graph: graphSummary(f), Numerator: numerator, Denominator: len(results) * relationDenominator, ConformanceDecision: scenarioDecision(decisions), SourceResolution: "EXACT", LineageResolution: resolution(lineageExact), Decisions: decisions, Transitions: transitions, Metrics: results}
}

type evaluator struct {
	fixture                      fixture
	byID                         map[string]metricSpec
	memo                         map[string]MetricResult
	visiting                     map[string]bool
	sourceDigest, semanticDigest string
}

func (e *evaluator) run(id string) MetricResult {
	if result, ok := e.memo[id]; ok {
		return result
	}
	metric := e.byID[id]
	if e.visiting[id] {
		return metricResult(metric, e.fixture.Artifacts[id], "UNKNOWN", dependencyIssue("DEPENDENCY_CYCLE", id), e.sourceDigest, e.semanticDigest, e.fixture)
	}
	e.visiting[id] = true
	result := e.directResult(metric)
	for _, dependency := range metric.DependsOn {
		upstream, ok := e.byID[dependency]
		if !ok {
			result = metricResult(metric, e.fixture.Artifacts[id], "UNKNOWN", dependencyIssue("UPSTREAM_METRIC_MISSING", dependency), e.sourceDigest, e.semanticDigest, e.fixture)
			break
		}
		if upstreamResult := e.run(upstream.ID); upstreamResult.Decision != "PASS" {
			result = metricResult(metric, e.fixture.Artifacts[id], "UNKNOWN", dependencyIssue("UPSTREAM_"+upstreamResult.Decision, dependency), e.sourceDigest, e.semanticDigest, e.fixture)
			result.LineageResolution = "LOWER_RESOLUTION"
			break
		}
	}
	e.visiting[id] = false
	e.memo[id] = result
	return result
}

func (e *evaluator) directResult(metric metricSpec) MetricResult {
	observations := e.fixture.Artifacts[metric.ID]
	for _, observation := range observations {
		if observation.Relation == "CONSUMES" && observation.RelationStatus == "REMOVED" {
			issue := &Issue{Stage: "SCENARIO", Step: "verify-counterexample", Reason: "EXPLICIT_CONSUMER_LINEAGE_COUNTEREXAMPLE", Cause: "VERIFIED_CONTRADICTION"}
			return metricResult(metric, observations, "FAIL_CLOSED", issue, e.sourceDigest, e.semanticDigest, e.fixture)
		}
	}
	return metricResult(metric, observations, "", nil, e.sourceDigest, e.semanticDigest, e.fixture)
}
