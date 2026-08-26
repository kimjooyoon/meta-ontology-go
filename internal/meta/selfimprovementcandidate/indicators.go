package selfimprovementcandidate

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type indicatorDefinition struct{ id, class, proof, operation string }

var indicatorDefinitions = []indicatorDefinition{
	{"foundation.source-schema", "DRIVER", "FOUNDATION", "bind-observation-schema"},
	{"foundation.exact-head", "DRIVER", "FOUNDATION", "bind-exact-observation-head"},
	{"foundation.source-digest", "DRIVER", "FOUNDATION", "verify-observation-digest"},
	{"foundation.fixed-source", "DRIVER", "FOUNDATION", "bind-observation-denominators"},
	{"foundation.gooo-contract", "DRIVER", "FOUNDATION", "compile-candidate-contract"},
	{"foundation.policy-version", "DRIVER", "FOUNDATION", "bind-candidate-policy"},
	{"coherence.explicit-gap", "OUTCOME", "COHERENCE", "select-explicit-nonclaim"},
	{"coherence.deterministic-selection", "DRIVER", "COHERENCE", "apply-fixed-gap-priority"},
	{"coherence.before-coordinate", "OUTCOME", "COHERENCE", "bind-zero-witness-baseline"},
	{"coherence.target-coordinate", "OUTCOME", "COHERENCE", "bind-one-witness-target"},
	{"coherence.candidate-count", "OUTCOME", "COHERENCE", "emit-one-data-candidate"},
	{"coherence.reader-views", "DRIVER", "COHERENCE", "project-candidate-resolutions"},
	{"regression.source-candidate-zero", "GUARDRAIL", "REGRESSION", "deny-upstream-candidate-leak"},
	{"regression.source-authority-zero", "GUARDRAIL", "REGRESSION", "deny-upstream-effect-leak"},
	{"regression.execution-authority-zero", "GUARDRAIL", "REGRESSION", "deny-candidate-execution"},
	{"regression.repository-writes-zero", "GUARDRAIL", "REGRESSION", "deny-repository-writes"},
}

func buildIndicators(success bool) []Indicator {
	value := 0
	if success {
		value = 1
	}
	result := make([]Indicator, 0, len(indicatorDefinitions))
	for _, definition := range indicatorDefinitions {
		result = append(result, Indicator{ID: definition.id, Class: definition.class,
			ProofChoice: definition.proof, MetaOperation: definition.operation,
			Value: value, Target: 1, Satisfied: success})
	}
	return result
}
