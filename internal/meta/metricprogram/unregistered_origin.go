package metricprogram

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

func unregisteredBindingError(plan StrategyPlan, binding StrategyBinding, inputs unresolvedpolicy.InputDigests) error {
	if err := VerifyCompiledRegistry(); err != nil {
		return fmt.Errorf("indicator %q operation %q is not resolved: %w", binding.IndicatorID, binding.MetaOperation, err)
	}
	registry := RegistrySourceOrigin()
	origin := unresolvedpolicy.OperationOrigin{
		Repository: plan.Repository, SubjectSHA: plan.SubjectSHA, Inputs: inputs,
		IndicatorID: binding.IndicatorID, RequestedOperation: binding.MetaOperation,
		Registry: unresolvedpolicy.ProjectionOrigin{
			SourcePath: registry.SourcePath, SourceDigest: registry.SourceDigest,
			ProjectionPath: registry.ProjectionPath, ProjectionDigest: registry.ProjectionDigest,
		},
	}
	refusal, err := unresolvedpolicy.NewOperationRefusal(origin)
	if err != nil {
		return fmt.Errorf("indicator %q operation %q is not resolved: %w", binding.IndicatorID, binding.MetaOperation, err)
	}
	return fmt.Errorf("indicator %q is not resolved: %w", binding.IndicatorID, refusal)
}
