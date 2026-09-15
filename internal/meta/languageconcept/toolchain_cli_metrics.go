package languageconcept

func toolchainCLIMetricBindings() []string {
	return []string{
		"gooo.metric.toolchain.cli-readiness-bps.v1",
		"gooo.metric.toolchain.cli-positive-paths.v1",
		"gooo.metric.toolchain.cli-guardrail-rejections.v1",
		"gooo.metric.toolchain.cli-executed-cases.v1",
		"gooo.metric.toolchain.cli-invocations.v1",
		"gooo.metric.toolchain.cli-declared-commands.v1",
		"gooo.metric.toolchain.cli-structured-outputs.v1",
		"gooo.metric.toolchain.cli-language-operations.v1",
		"gooo.metric.toolchain.cli-deterministic-replays.v1",
		"gooo.metric.toolchain.cli-binary-bindings.v1",
		"gooo.metric.toolchain.cli-unresolved.guardrail.v1",
		"gooo.metric.toolchain.cli-exit-mismatch.guardrail.v1",
		"gooo.metric.toolchain.cli-stdout-mismatch.guardrail.v1",
		"gooo.metric.toolchain.cli-stderr-mismatch.guardrail.v1",
		"gooo.metric.toolchain.cli-replay-mismatch.guardrail.v1",
		"gooo.metric.toolchain.cli-repository-writes.guardrail.v1",
		"gooo.metric.toolchain.cli-mutation-authority.guardrail.v1",
		"gooo.metric.toolchain.cli-registry-drift.guardrail.v1",
	}
}
