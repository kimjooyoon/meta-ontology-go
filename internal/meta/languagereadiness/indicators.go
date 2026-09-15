package languagereadiness

func buildIndicators(snapshot Snapshot) []Indicator {
	summary := snapshot.Summary
	indicators := []Indicator{
		indicator("gooo.metric.language.self-improvement-readiness-bps.v1", "outcome", "COHERENCE", "quantify-closed-obligation-set", summary.ReadinessBPS, 10000),
		indicator("gooo.metric.language.completed-obligations.v1", "outcome", "FOUNDATION", "count-explicit-conformance", summary.Completed, summary.Total),
	}
	for _, area := range []struct {
		name, metric, proof string
	}{
		{areaLanguage, "gooo.metric.language.area.language-completed.v1", "FOUNDATION"},
		{areaToolchain, "gooo.metric.language.area.toolchain-completed.v1", "COHERENCE"},
		{areaMeta, "gooo.metric.language.area.meta-completed.v1", "COHERENCE"},
		{areaAutonomy, "gooo.metric.language.area.autonomy-completed.v1", "REGRESSION"},
	} {
		completed, total := areaCounts(snapshot.Obligations, area.name)
		indicators = append(indicators, indicator(area.metric, "driver", area.proof,
			"count-area-conformance", completed, total))
	}
	indicators = append(indicators,
		indicator("gooo.metric.language.unresolved-obligations.guardrail.v1", "guardrail", "COHERENCE", "lower-readiness-resolution", summary.Unresolved, 0),
		indicator("gooo.metric.language.observer-writes.guardrail.v1", "guardrail", "FOUNDATION", "preserve-read-only-readiness", snapshot.RepositoryWrites, 0),
	)
	return indicators
}

func indicator(id, class, proof, operation string, value, target int) Indicator {
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof,
		Producer: "languagereadiness.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: operation, Value: value, Target: target, Satisfied: value == target,
	}
}

func areaCounts(results []ObligationResult, area string) (int, int) {
	completed, total := 0, 0
	for _, result := range results {
		if result.Area != area {
			continue
		}
		total++
		if result.Status == "SATISFIED" {
			completed++
		}
	}
	return completed, total
}
