package valueexecution

func buildIndicators(checks evidence) []Indicator {
	definitions := []struct {
		id, class, operation string
		value                bool
	}{
		{"source-parsed", "FOUNDATION", "parse-gooo-value-source", checks.sourceParsed},
		{"activity-resolved", "FOUNDATION", "resolve-gooo-activity", checks.activityResolved},
		{"program-present", "FOUNDATION", "read-computes-clause", checks.programPresent},
		{"program-semantic-binding", "COHERENCE", "bind-program-to-bidir-activity", checks.semanticBound},
		{"fingerprint-sensitive", "COHERENCE", "vary-program-and-compare-semantic-fingerprint", checks.fingerprintSensitive},
		{"registry-operation-known", "COHERENCE", "resolve-operation-registry-entry", checks.registryKnown},
		{"operand-parsed", "COHERENCE", "compile-operation-operand", checks.operandParsed},
		{"signature-supported", "COHERENCE", "check-value-signature", checks.signatureSupported},
		{"value-cases-exact", "REGRESSION", "execute-fixed-value-cases", checks.valueCasesExact},
		{"outputs-observed", "REGRESSION", "record-gooo-value-outputs", checks.outputsObserved},
		{"deterministic-replay", "REGRESSION", "replay-value-cases", checks.deterministicReplay},
		{"counterexamples-exact", "REGRESSION", "execute-fixed-counterexamples", checks.counterexamplesExact},
		{"unknown-operation-fail-closed", "REGRESSION", "reject-unknown-operation", checks.unknownFailClosed},
		{"malformed-operand-fail-closed", "REGRESSION", "reject-malformed-operand", checks.operandFailClosed},
		{"overflow-fail-closed", "REGRESSION", "reject-int64-overflow", checks.overflowFailClosed},
		{"core-ir-program-preserved", "COHERENCE", "lower-program-into-core-ir", checks.coreIRProgramPreserved},
		{"core-ir-fingerprint-sensitive", "COHERENCE", "vary-program-and-compare-core-ir", checks.coreIRFingerprintSensitive},
		{"core-ir-unknown-attribute-fail-closed", "REGRESSION", "reject-unknown-core-ir-attribute", checks.coreIRUnknownAttributeFailClosed},
	}
	indicators := make([]Indicator, 0, len(definitions))
	for _, definition := range definitions {
		indicators = append(indicators, Indicator{
			ID: definition.id, Class: definition.class, ProofChoice: definition.class,
			MetaOperation: definition.operation, Value: boolInt(definition.value), Target: 1, Satisfied: definition.value,
		})
	}
	return indicators
}
