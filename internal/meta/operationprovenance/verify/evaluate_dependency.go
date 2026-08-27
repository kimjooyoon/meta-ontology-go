package verify

type evaluator struct {
	fixture                      cFixture
	byID                         map[string]cMetric
	memo                         map[string]metricResult
	visiting                     map[string]bool
	sourceDigest, semanticDigest string
}

func newEvaluator(f cFixture, sourceDigest, semanticDigest string) *evaluator {
	byID := map[string]cMetric{}
	for _, metric := range f.metrics {
		byID[metric.id] = metric
	}
	return &evaluator{fixture: f, byID: byID, memo: map[string]metricResult{}, visiting: map[string]bool{}, sourceDigest: sourceDigest, semanticDigest: semanticDigest}
}

func (e *evaluator) run(id string) metricResult {
	if result, ok := e.memo[id]; ok {
		return result
	}
	metric := e.byID[id]
	if e.visiting[id] {
		return resultFor(metric, e.fixture.artifacts[id], "UNKNOWN", dependencyIssue("DEPENDENCY_CYCLE", id), e.sourceDigest, e.semanticDigest, e.fixture)
	}
	e.visiting[id] = true
	result := e.directResult(metric)
	for _, dependency := range metric.dependsOn {
		if _, ok := e.byID[dependency]; !ok {
			result = resultFor(metric, e.fixture.artifacts[id], "UNKNOWN", dependencyIssue("UPSTREAM_METRIC_MISSING", dependency), e.sourceDigest, e.semanticDigest, e.fixture)
			break
		}
		if upstream := e.run(dependency); upstream.Decision != "PASS" {
			result = resultFor(metric, e.fixture.artifacts[id], "UNKNOWN", dependencyIssue("UPSTREAM_"+upstream.Decision, dependency), e.sourceDigest, e.semanticDigest, e.fixture)
			result.LineageResolution = "LOWER_RESOLUTION"
			break
		}
	}
	e.visiting[id] = false
	e.memo[id] = result
	return result
}

func (e *evaluator) directResult(metric cMetric) metricResult {
	observations := e.fixture.artifacts[metric.id]
	for _, observation := range observations {
		if observation.Relation == "CONSUMES" && observation.RelationStatus == "REMOVED" {
			issue := &issue{Stage: "SCENARIO", Step: "verify-counterexample", Reason: "EXPLICIT_CONSUMER_LINEAGE_COUNTEREXAMPLE", Cause: "VERIFIED_CONTRADICTION"}
			return resultFor(metric, observations, "FAIL_CLOSED", issue, e.sourceDigest, e.semanticDigest, e.fixture)
		}
	}
	return resultFor(metric, observations, "", nil, e.sourceDigest, e.semanticDigest, e.fixture)
}

func dependencyIssue(reason, blocked string) *issue {
	return &issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: reason, Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{blocked}}
}
