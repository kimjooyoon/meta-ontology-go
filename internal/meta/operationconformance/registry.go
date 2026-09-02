package operationconformance

var fixedIndicators = []IndicatorDefinition{
	{"filesystem.atomic-replacement/v1", "GUARDRAIL", "COHERENCE", "atomic-replacement-v1"},
	{"go.filename.build-semantics/v1", "DRIVER", "FOUNDATION", "build-selected-file-set-v1"},
	{"go.header.preserved/v1", "GUARDRAIL", "REGRESSION", "header-byte-digest-v1"},
	{"go.import.identity/v1", "OUTCOME", "FOUNDATION", "import-alias-path-set-union-v1"},
	{"go.initialization.order/v1", "OUTCOME", "FOUNDATION", "initialization-units-order-v1"},
	{"go.package.conformance/v1", "OUTCOME", "FOUNDATION", "selected-package-set-v1"},
}

type oracleIdentity struct {
	ID, Indicator string
	Expected      Decision
}

var fixedOracleCases = []oracleIdentity{
	{"atomic-replacement/pass", fixedIndicators[0].ID, DecisionPass},
	{"atomic-replacement/fail", fixedIndicators[0].ID, DecisionFail},
	{"atomic-replacement/unknown", fixedIndicators[0].ID, DecisionUnknown},
	{"build-semantics/pass", fixedIndicators[1].ID, DecisionPass},
	{"build-semantics/fail", fixedIndicators[1].ID, DecisionFail},
	{"build-semantics/unknown", fixedIndicators[1].ID, DecisionUnknown},
	{"header-preserved/pass", fixedIndicators[2].ID, DecisionPass},
	{"header-preserved/fail", fixedIndicators[2].ID, DecisionFail},
	{"header-preserved/unknown", fixedIndicators[2].ID, DecisionUnknown},
	{"import-identity/pass", fixedIndicators[3].ID, DecisionPass},
	{"import-identity/fail", fixedIndicators[3].ID, DecisionFail},
	{"import-identity/unknown", fixedIndicators[3].ID, DecisionUnknown},
	{"initialization-order/pass", fixedIndicators[4].ID, DecisionPass},
	{"initialization-order/fail", fixedIndicators[4].ID, DecisionFail},
	{"initialization-order/unknown", fixedIndicators[4].ID, DecisionUnknown},
	{"package-conformance/pass", fixedIndicators[5].ID, DecisionPass},
	{"package-conformance/fail", fixedIndicators[5].ID, DecisionFail},
	{"package-conformance/unknown", fixedIndicators[5].ID, DecisionUnknown},
}
