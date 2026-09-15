package metainvocation

import "fmt"

const (
	operationGoRule   = "ci.rule:go:v1"
	operationDocsRule = "ci.rule:docs:v1"
	operationYAMLRule = "ci.rule:yaml:v1"
	operationPlan     = "ci.plan:v1"
	operationVerify   = "ci.verify:v1"
)

type Registry struct {
	operations map[string]OperationSpec
}

func NewRegistry(specs ...OperationSpec) (Registry, error) {
	registry := Registry{operations: make(map[string]OperationSpec, len(specs))}
	for _, spec := range specs {
		if spec.ID == "" || spec.InputEntity == "" || spec.OutputEntity == "" || spec.Phase == "" {
			return Registry{}, fmt.Errorf("operation spec is incomplete")
		}
		if spec.RepositoryWrites != 0 || spec.MutationAuthority {
			return Registry{}, fmt.Errorf("operation %q exceeds the zero-effect meta-invocation boundary", spec.ID)
		}
		if _, exists := registry.operations[spec.ID]; exists {
			return Registry{}, fmt.Errorf("operation %q is registered twice", spec.ID)
		}
		registry.operations[spec.ID] = spec
	}
	return registry, nil
}

func StandardRegistry() Registry {
	registry, err := NewRegistry(
		OperationSpec{ID: operationGoRule, InputEntity: "ChangeSet", OutputEntity: "CheckPlan", Phase: "RULE_SELECTION"},
		OperationSpec{ID: operationDocsRule, InputEntity: "ChangeSet", OutputEntity: "CheckPlan", Phase: "RULE_SELECTION"},
		OperationSpec{ID: operationYAMLRule, InputEntity: "ChangeSet", OutputEntity: "CheckPlan", Phase: "RULE_SELECTION"},
		OperationSpec{ID: operationPlan, InputEntity: "ChangeSet", OutputEntity: "CheckPlan", Phase: "PLAN_COMPOSITION"},
		OperationSpec{ID: operationVerify, InputEntity: "CheckPlan", OutputEntity: "VerificationReceipt", Phase: "VERIFICATION"},
	)
	if err != nil {
		panic(err)
	}
	return registry
}

func (r Registry) lookup(id string) (OperationSpec, bool) {
	spec, ok := r.operations[id]
	return spec, ok
}
