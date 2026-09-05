package transformationeffect

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

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
	progressSequence := 0
	for _, action := range actions {
		before, err := workspace.Scan(box.Root)
		if err != nil {
			return result, err
		}
		if failure, found := inputFailureForAction(in.receipts.Failures, action); found {
			result.effects = append(result.effects, effectForFailure(action, before, failure))
			continue
		}
		preflight, err := runActionWithProgress(box, opts, in.plan, action, true, &progressSequence)
		if err != nil {
			return result, err
		}
		applied, err := runActionWithProgress(box, opts, in.plan, action, false, &progressSequence)
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
		receipt = preserveInputInstanceEvidence(receipt, in.receipts.Receipts)
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

func runActionWithProgress(box *workspace.Sandbox, opts Options, plan generation.Plan, action generation.Action, check bool, sequence *int) ([]byte, error) {
	phase := "APPLY"
	if check {
		phase = "PREFLIGHT"
	}
	return runProgressPhase(opts, action, phase, sequence, func() ([]byte, error) {
		return runAction(box, opts, plan, action, check)
	})
}

func runProgressPhase(opts Options, action generation.Action, phase string, sequence *int, execute func() ([]byte, error)) ([]byte, error) {
	if err := writeOperationProgress(opts, action, phase, "ENTERED", "", sequence); err != nil {
		warnOperationProgress(opts, action, phase, "ENTERED", err)
	}
	output, runErr := execute()
	returnError := ""
	if runErr != nil {
		returnError = "ERROR"
	}
	if err := writeOperationProgress(opts, action, phase, "RETURNED", returnError, sequence); err != nil {
		warnOperationProgress(opts, action, phase, "RETURNED", err)
	}
	return output, runErr
}

func warnOperationProgress(opts Options, action generation.Action, phase, boundary string, err error) {
	fmt.Fprintf(os.Stderr, "transformation-effect: operation progress unavailable invocation=%s action_indicator_id=%s operation=%s phase=%s boundary=%s: %v\n",
		opts.InvocationID, action.IndicatorID, action.Operation, phase, boundary, err)
}

func writeOperationProgress(opts Options, action generation.Action, phase, boundary, returnError string, sequence *int) error {
	if opts.ProgressWriter == nil {
		return nil
	}
	*sequence = *sequence + 1
	event := operationProgressEvent{
		Schema:                      "gooo/transformation-effect-operation-progress/v1",
		HeadSHA:                     opts.ExpectedSHA,
		InvocationID:                opts.InvocationID,
		Sequence:                    *sequence,
		ActionIndicatorID:           action.IndicatorID,
		Operation:                   string(action.Operation),
		Activity:                    action.Activity,
		Executor:                    action.Executor,
		Subject:                     action.Subject,
		SubjectKind:                 string(action.SubjectKind),
		InputContractSourceDigest:   action.InputContractSourceDigest,
		InputContractSemanticDigest: action.InputContractSemanticDigest,
		Phase:                       phase,
		Boundary:                    boundary,
		ReturnError:                 returnError,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if written, err := opts.ProgressWriter.Write(append(payload, '\n')); err != nil {
		return err
	} else if written != len(payload)+1 {
		return fmt.Errorf("short operation progress write: wrote %d of %d bytes", written, len(payload)+1)
	}
	return nil
}

func preserveInputInstanceEvidence(receipt generation.OperationReceipt, inputs []generation.OperationReceipt) generation.OperationReceipt {
	for _, input := range inputs {
		if input.ActionIndicatorID != receipt.ActionIndicatorID || input.InstanceEvidence == nil {
			continue
		}
		receipt.Indicators = append([]generation.IndicatorReceipt{}, input.Indicators...)
		evidence := *input.InstanceEvidence
		evidence.EvidenceOrigin = generation.EvidenceOriginInputReceipt
		evidence.SourceReceiptDigest = input.ReceiptDigest
		if !strings.HasPrefix(evidence.SourceReceiptDigest, "sha256:") {
			evidence.SourceReceiptDigest = "sha256:" + evidence.SourceReceiptDigest
		}
		return generation.AttachInstanceEvidence(receipt, evidence)
	}
	return receipt
}

func inputFailureForAction(failures []generation.ObservationFailure, action generation.Action) (generation.ObservationFailure, bool) {
	for _, failure := range failures {
		if failure.ActionIndicatorID == action.IndicatorID {
			return failure, true
		}
	}
	return generation.ObservationFailure{}, false
}
