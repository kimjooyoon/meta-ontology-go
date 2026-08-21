package transformationeffect

import (
	"errors"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type executionResult struct {
	effects    []Effect
	receipts   generation.ReceiptReport
	provenance generation.ArtifactProvenance
	baseline   treeState
	final      treeState
	patch      Patch
}

func executePlan(in inputSet, opts Options, source treeState) (result executionResult, err error) {
	if in.plan.Decision != generation.DecisionFixedPoint && in.plan.Decision != generation.DecisionPlan {
		return result, fmt.Errorf("generation decision %s is not executable", in.plan.Decision)
	}
	box, baseline, err := openSandbox(opts.Root, opts.ExpectedSHA, source)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, box.close()) }()
	result.baseline = baseline
	actions := append([]generation.Action{}, in.plan.Selected...)
	sort.Slice(actions, func(i, j int) bool { return actions[i].IndicatorID < actions[j].IndicatorID })
	sealed := make([]generation.OperationReceipt, 0, len(actions))
	for _, action := range actions {
		before, err := scanTree(box.root)
		if err != nil {
			return result, err
		}
		preflight, err := runAction(box, opts, in.plan, action, true)
		if err != nil {
			return result, err
		}
		applied, err := runAction(box, opts, in.plan, action, false)
		if err != nil {
			return result, err
		}
		after, err := scanTree(box.root)
		if err != nil {
			return result, err
		}
		changes := makePatch(opts.ExpectedSHA, before, after).Changes
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
		observations := make([]generation.IndicatorReceipt, 0, len(action.RequiredIndicatorIDs))
		for _, id := range action.RequiredIndicatorIDs {
			observations = append(observations, generation.IndicatorReceipt{ID: id,
				Verdict: generation.IndicatorVerdictPass, EvidenceDigest: hashJSON([]string{evidence, id}), ProofChoice: action.ProofChoice})
		}
		receipt := generation.SealReceipt(in.plan, action, observations)
		sealed = append(sealed, receipt)
		result.effects = append(result.effects, effectFor(action, before, after, changes, evidence, receipt.ReceiptDigest))
	}
	result.final, err = scanTree(box.root)
	if err != nil {
		return result, err
	}
	result.patch = makePatch(opts.ExpectedSHA, baseline, result.final)
	result.receipts = generation.VerifyReceipts(in.plan, sealed)
	result.provenance = generation.BindArtifactProvenance(in.plan, in.execution, result.receipts)
	if result.provenance.Decision != generation.ArtifactProvenanceDecisionBound {
		return result, fmt.Errorf("executed provenance is not bound")
	}
	return result, nil
}

func effectFor(action generation.Action, before, after treeState, changes []PatchChange, evidence, receipt string) Effect {
	return Effect{ActionIndicatorID: action.IndicatorID, MetricID: string(action.MetricID), Subject: action.Subject,
		SubjectKind: string(action.SubjectKind), Operation: string(action.Operation), Executor: action.Executor,
		Evaluator: action.Evaluator, ProofChoice: string(action.ProofChoice), BeforeTreeDigest: before.Digest,
		AfterTreeDigest: after.Digest, ChangedPathCount: len(changes), ChangedPathDigest: hashJSON(changes),
		ResidualActionable: 0, EvaluatorEvidence: evidence, ReceiptDigest: receipt, Status: "APPLIED"}
}
