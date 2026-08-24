package languageassurance

func writeSetIndicator(summary Summary) Indicator {
	value := summary.RepositoryWrites
	result := indicator("gooo.metric.effects.write-set-exactness.v1", ClassGuardrail, ProofRegression, "observe-exact-write-set", &value, 0, "paths", RelationLessOrEqual)
	result.Producer = "internal/meta/writeset.Compare"
	result.Consumer = "language-assurance-gate"
	return result
}
