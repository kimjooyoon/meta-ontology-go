package verify

type evaluator struct {
	fixture                      cFixture
	edges                        map[string]int
	sourceDigest, semanticDigest string
	byID                         map[string]cMetric
	memo                         map[string]metricResult
	visiting                     map[string]bool
}

func newEvaluator(f cFixture, edges map[string]int, sourceDigest, semanticDigest string) *evaluator {
	byID := make(map[string]cMetric, len(f.metrics))
	for _, metric := range f.metrics {
		byID[metric.id] = metric
	}
	return &evaluator{fixture: f, edges: edges, sourceDigest: sourceDigest, semanticDigest: semanticDigest, byID: byID, memo: map[string]metricResult{}, visiting: map[string]bool{}}
}

func (e *evaluator) run(id string) metricResult {
	if result, ok := e.memo[id]; ok {
		return result
	}
	metric := e.byID[id]
	if e.visiting[id] {
		return resultFor(metric, e.edges, "UNKNOWN", dependencyIssue("DEPENDENCY_CYCLE", id), e.sourceDigest, e.semanticDigest, e.fixture)
	}
	e.visiting[id] = true
	result := resultFor(metric, e.edges, "", nil, e.sourceDigest, e.semanticDigest, e.fixture)
	for _, dependency := range metric.dependsOn {
		if _, ok := e.byID[dependency]; !ok {
			result = resultFor(metric, e.edges, "UNKNOWN", dependencyIssue("UPSTREAM_METRIC_MISSING", dependency), e.sourceDigest, e.semanticDigest, e.fixture)
			break
		}
		if upstream := e.run(dependency); upstream.Decision != "PASS" {
			result = resultFor(metric, e.edges, "UNKNOWN", dependencyIssue("UPSTREAM_"+upstream.Decision, dependency), e.sourceDigest, e.semanticDigest, e.fixture)
			break
		}
	}
	e.visiting[id] = false
	e.memo[id] = result
	return result
}

func dependencyIssue(reason, blocked string) *issue {
	return &issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: reason, Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{blocked}}
}
