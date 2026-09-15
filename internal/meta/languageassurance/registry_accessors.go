package languageassurance

func Denominator() []ObligationDefinition {
	return append([]ObligationDefinition(nil), denominatorV1...)
}

func RoleConflictPairs() []RolePair { return append([]RolePair(nil), conflictPairs...) }

func UnknownLaunderingOutputs() []Decision { return append([]Decision(nil), launderingOutputs...) }

func obligation(metricID string, priority Priority, class IndicatorClass, proof ProofChoice, operation string) ObligationDefinition {
	return ObligationDefinition{MetricID: metricID, Priority: priority, Class: class, ProofChoice: proof, RequiredMetaOperation: operation}
}
