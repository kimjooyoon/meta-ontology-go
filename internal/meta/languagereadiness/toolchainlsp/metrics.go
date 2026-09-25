package toolchainlsp

const metricPrefix = "gooo.metric.toolchain.lsp-"

var outcomeMetrics = []string{
	"readiness-bps.v1", "case-readiness-bps.v1", "proof-readiness-bps.v1",
}

var driverMetrics = []string{
	"executed-cases.v1", "protocol-cases.v1", "coupling-cases.v1",
	"advertised-capabilities.v1", "read-features.v1", "diagnostic-paths.v1",
	"navigation-paths.v1", "symbol-paths.v1", "semantic-token-paths.v1",
	"utf16-replays.v1", "transcript-replays.v1", "fail-closed-paths.v1",
	"concept-bindings.v1", "code-bindings.v1", "metric-bindings.v1",
	"use-case-bindings.v1",
}

var guardrailMetrics = []string{
	"missing-cases.guardrail.v1", "unexpected-cases.guardrail.v1",
	"case-failures.guardrail.v1", "capability-gaps.guardrail.v1",
	"unexpected-protocol-errors.guardrail.v1", "diagnostic-gaps.guardrail.v1",
	"nonstandard-wire-fields.guardrail.v1", "stale-navigation-leaks.guardrail.v1",
	"unknown-navigation-leaks.guardrail.v1", "fail-closed-navigation-leaks.guardrail.v1",
	"unresolved.guardrail.v1", "digest-failures.guardrail.v1",
	"corpus-drift.guardrail.v1", "concept-drift.guardrail.v1",
	"head-mismatches.guardrail.v1", "proof-failures.guardrail.v1",
	"repository-writes.guardrail.v1", "mutation-authorities.guardrail.v1",
}

func MetricIDs() []string {
	all := append(append([]string{}, outcomeMetrics...), driverMetrics...)
	all = append(all, guardrailMetrics...)
	for index := range all {
		all[index] = metricPrefix + all[index]
	}
	return all
}
