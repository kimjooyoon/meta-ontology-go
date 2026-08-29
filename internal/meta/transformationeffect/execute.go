package transformationeffect

import (
	"errors"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect/workspace"
)

func executePlan(in inputSet, opts Options, source workspace.State) (result executionResult, err error) {
	if in.plan.Decision != generation.DecisionFixedPoint && in.plan.Decision != generation.DecisionPlan {
		return result, fmt.Errorf("generation decision %s is not executable", in.plan.Decision)
	}
	box, baseline, err := workspace.Open(opts.Root, opts.ExpectedSHA, source)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, box.Close()) }()
	result.baseline = baseline
	actions := append([]generation.Action{}, in.plan.Selected...)
	sort.Slice(actions, func(i, j int) bool { return actions[i].IndicatorID < actions[j].IndicatorID })
	result.selectedPlanOperations = len(actions)
	for _, action := range actions {
		if _, err := resolveActionBinding(in.plan, action); err != nil {
			result.unboundExecutorOperations = len(actions) - result.boundExecutorOperations
			return result, err
		}
		result.boundExecutorOperations++
	}
	sealed := make([]generation.OperationReceipt, 0, len(actions))
	result.failures = append(result.failures, in.receipts.Failures...)
	for _, action := range actions {
		before, err := workspace.Scan(box.Root)
		if err != nil {
			return result, err
		}
		if failure, found := inputFailureForAction(in.receipts.Failures, action); found {
			result.effects = append(result.effects, effectForFailure(action, before, failure))
			continue
		}
		preflight, err := runAction(box, opts, in.plan, action, true)
		if err != nil {
			return result, err
		}
		applied, err := runAction(box, opts, in.plan, action, false)
		if err != nil {
			return result, err
		}
		after, err := workspace.Scan(box.Root)
		if err != nil {
			return result, err
		}
		changes := workspace.MakePatch(opts.ExpectedSHA, before, after).Changes
		metrics, metricPayload, err := freshMetrics(box, opts.ExpectedSHA)
		if err != nil || len(changes) == 0 {
			return result, fmt.Errorf("operation %s produced no verified effect: %w", action.Operation, err)
		}
		residual := residualActionable(metrics, action)
		if residual != 0 {
			return result, fmt.Errorf("operation %s left %d actionable subjects", action.Operation, residual)
		}
		evidence := hashJSON([]any{action.IndicatorID, hashBytes(preflight), hashBytes(applied),
			before.Digest, after.Digest, hashJSON(changes), hashBytes(metricPayload), residual})
		observations, splitGoEvaluation, evidence, err := operationObservations(box.Root, action, applied, evidence)
		if err != nil {
			return result, err
		}
		receipt := generation.SealReceipt(in.plan, action, observations)
		sealed = append(sealed, receipt)
		effect := effectFor(action, before, after, changes, evidence, receipt.ReceiptDigest)
		effect.SplitGoEvaluation = splitGoEvaluation
		result.effects = append(result.effects, effect)
	}
	result.final, err = workspace.Scan(box.Root)
	if err != nil {
		return result, err
	}
	result.patch = workspace.MakePatch(opts.ExpectedSHA, baseline, result.final)
	result.receipts = generation.VerifyReceiptsWithFailures(in.plan, sealed, result.failures)
	result.provenance = generation.BindArtifactProvenance(in.plan, in.execution, result.receipts)
	if result.provenance.Decision != generation.ArtifactProvenanceDecisionBound {
		return result, fmt.Errorf("executed provenance is not bound")
	}
	return result, nil
}

func inputFailureForAction(failures []generation.ObservationFailure, action generation.Action) (generation.ObservationFailure, bool) {
	for _, failure := range failures {
		if failure.ActionIndicatorID == action.IndicatorID {
			return failure, true
		}
	}
	return generation.ObservationFailure{}, false
}
