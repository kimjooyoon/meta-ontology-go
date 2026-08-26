package valuecatalog

func buildIndicators(report Report) []Indicator {
	definitions := []struct {
		id, class, proof, operation string
		value                       bool
	}{
		{"source-parsed", "DRIVER", "FOUNDATION", "parse-catalog-source", report.SourceDigest != ""},
		{"activities-observed", "DRIVER", "FOUNDATION", "bind-two-catalog-activities", report.ActivitiesObserved == 2},
		{"baseline-program-bound", "DRIVER", "COHERENCE", "compile-baseline-program", report.Baseline.Program == "int.add:1"},
		{"baseline-core-preserved", "DRIVER", "COHERENCE", "lower-baseline-program", report.BaselineCoreProgram == "int.add:1"},
		{"baseline-cases-exact", "DRIVER", "REGRESSION", "execute-baseline-cases", report.Baseline.Passed == 3},
		{"extension-before-missing", "DRIVER", "FOUNDATION", "remove-extension-program", report.Improvement.BeforeEvidence == "VALUE_PROGRAM_MISSING"},
		{"extension-program-exact", "OUTCOME", "COHERENCE", "compile-source-only-extension", report.Extension.Program == "int.add:2"},
		{"extension-core-preserved", "OUTCOME", "COHERENCE", "lower-source-only-extension", report.ExtensionCoreProgram == "int.add:2"},
		{"core-fingerprint-sensitive", "OUTCOME", "COHERENCE", "compare-before-after-core-ir", report.Summary.CoreFingerprintSensitive.Satisfied == 1},
		{"extension-cases-exact", "OUTCOME", "REGRESSION", "execute-extension-cases", report.Extension.Passed == 3},
		{"unknown-operation-rejected", "GUARDRAIL", "REGRESSION", "reject-unknown-extension-operation", report.Summary.UnknownCounterexamplePassed},
		{"repository-writes-zero", "GUARDRAIL", "REGRESSION", "observe-without-repository-writes", report.Summary.RepositoryWrites == 0},
		{"authority-denied", "GUARDRAIL", "REGRESSION", "deny-automatic-authority", !report.Authority.RepositoryMutationAuthorized && !report.Authority.PromotionAuthorized && !report.Authority.AutomaticAdoptionAuthorized},
	}
	indicators := make([]Indicator, 0, len(definitions))
	for _, definition := range definitions {
		indicators = append(indicators, Indicator{
			ID: definition.id, Class: definition.class, ProofChoice: definition.proof,
			MetaOperation: definition.operation, Value: boolInt(definition.value), Target: 1, Satisfied: definition.value,
		})
	}
	return indicators
}
