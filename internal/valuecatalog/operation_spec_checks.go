package valuecatalog

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

type operationSpecCheck struct {
	id, class, proof, operation, stage, statement, evidence string
	satisfied                                               bool
}

func operationSpecChecks(report Report) []operationSpecCheck {
	var spec valueexecution.OperationSpec
	if len(report.OperationSpecs) == 1 {
		spec = report.OperationSpecs[0]
	}
	validSpec := valueexecution.ValidateOperationSpec(spec) == nil
	baselineBound := valueexecution.ValidateOperationIR(report.Baseline.Operation) == nil
	extensionBound := report.Extension.CompileReason == valueexecution.ReasonProgramMissing ||
		valueexecution.ValidateOperationIR(report.Extension.Operation) == nil
	evidence := digestValue(struct {
		Specs     []valueexecution.OperationSpec
		Baseline valueexecution.OperationIR
		Extension valueexecution.OperationIR
	}{report.OperationSpecs, report.Baseline.Operation, report.Extension.Operation})
	return []operationSpecCheck{
		{"catalog-singleton", "DRIVER", "FOUNDATION", "bind-closed-operation-catalog", "CATALOG/resolve-spec", "one canonical operation spec is resolved", len(report.OperationSpecs) == 1 && validSpec, evidence},
		{"identity-versioned", "DRIVER", "FOUNDATION", "bind-operation-identity", "RESOLVE/identity", "operation identity and version are explicit", spec.ID == "int.add" && spec.Version == 1, evidence},
		{"signature-typed", "DRIVER", "COHERENCE", "type-operation-signature", "TYPECHECK/signature", "input and output entities are typed", spec.Arity == 1 && slices.Equal(spec.InputEntities, []string{valueexecution.IntegerEntity}) && spec.OutputEntity == valueexecution.IntegerEntity, evidence},
		{"operand-typed", "DRIVER", "COHERENCE", "type-operation-operand", "TYPECHECK/operand", "the literal operand kind is explicit", spec.OperandKind == valueexecution.OperandInt64Literal, evidence},
		{"effect-explicit", "OUTCOME", "COHERENCE", "classify-operation-effect", "EFFECT/classify", "the operation effect is explicit", spec.Effect == valueexecution.EffectPureValue, evidence},
		{"determinism-explicit", "OUTCOME", "REGRESSION", "classify-operation-determinism", "EXECUTE/replay", "determinism is explicit", spec.Determinism == valueexecution.Deterministic, evidence},
		{"failure-set-closed", "GUARDRAIL", "REGRESSION", "bind-operation-failure-set", "EXECUTE/failure-set", "runtime failures form a closed set", slices.Equal(spec.FailureReasons, []string{valueexecution.ReasonInputArityMismatch, valueexecution.ReasonIntegerOverflow}), evidence},
		{"authority-zero", "GUARDRAIL", "REGRESSION", "deny-operation-authority", "AUTHORITY/deny", "repository, external, and promotion authority are absent", !spec.Authority.RepositoryWrite && !spec.Authority.ExternalCall && !spec.Authority.Promotion, evidence},
		{"invocation-ir-bound", "OUTCOME", "COHERENCE", "lower-typed-operation-invocation", "LOWER/invocation-ir", "every present invocation binds the canonical spec", baselineBound && extensionBound, evidence},
	}
}
