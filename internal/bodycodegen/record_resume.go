package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

const recordControlStageLimit = 16

// ResumeRecordAssembly verifies the saved generation, then applies a new Gooo
// policy to the remaining original ranking. Changed dependencies recheck the
// attempted prefix; the target contract and cumulative budget stay fixed.
// Historical model observations are retained; this function loads no model.
func ResumeRecordAssembly(ctx context.Context, filename string, priorSource, source []byte,
	prior Result, policy RecordAssemblyPolicy) (Result, error) {
	if ctx == nil || prior.Report.RecordAssembly == nil {
		return Result{}, fmt.Errorf("resume requires a record generation and context")
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(ctx, bodyPathBudget)
	defer cancel()
	if _, err := RealizeSourceAssembly(ctx, filename, priorSource, prior); err != nil {
		return Result{}, fmt.Errorf("resume original record: %w", err)
	}
	oldPlan, err := prepareRecordAssembly(ctx, filename, priorSource, prior.Report.Activity)
	if err != nil {
		return Result{}, err
	}
	plan, err := prepareRecordAssembly(ctx, filename, source, prior.Report.Activity)
	if err != nil {
		return Result{}, err
	}
	if err := compatibleRecordTarget(oldPlan, plan); err != nil {
		return Result{}, err
	}
	old := prior.Report.RecordAssembly
	r := newRecordAssemblyReceipt(source, plan)
	if len(old.ControlHistory) >= recordControlStageLimit {
		return Result{}, fmt.Errorf("record continuation supports at most %d saved stages", recordControlStageLimit)
	}
	copyRecordModelObservation(r, old)
	history := appendRecordHistory(old, priorSource, source)
	plans, origin, err := recordHistoryPlans(ctx, filename, source, plan, history)
	if err != nil {
		return Result{}, err
	}
	if err := verifyRecordAssemblyRanking(old, r, plans[0]); err != nil {
		return Result{}, err
	}
	r.RankingSourceSHA256 = origin
	current := &RecordAssemblyControl{Policy: policy}
	if err := replayRecordSearchStages(ctx, plans, r, history, current); err != nil {
		return Result{}, err
	}
	result, err := emitRecordAssembly(ctx, filename, source, plan, r)
	r.GenerationNS = time.Since(started).Nanoseconds()
	if err == nil {
		populateCompletenessReceipt(&result.Report, "")
	}
	return result, err
}

func copyRecordModelObservation(dst, src *RecordAssemblyReceipt) {
	dst.Ranking = append([]uint16(nil), src.Ranking...)
	dst.ModelRequested, dst.ModelCalls, dst.PredictNS = src.ModelRequested, src.ModelCalls, src.PredictNS
	if src.Model != nil {
		value := *src.Model
		dst.Model = &value
	}
	if src.Prediction != nil {
		value := *src.Prediction
		dst.Prediction = &value
	}
	if src.Context != nil {
		value := *src.Context
		value.Parts = append([]string(nil), src.Context.Parts...)
		dst.Context = &value
	}
}

func replayRecordSearchStages(ctx context.Context, plans []recordAssemblyPlan, r *RecordAssemblyReceipt,
	history []RecordAssemblyControlStage, current *RecordAssemblyControl) error {
	if len(history) > recordControlStageLimit || (len(history) > 0 && current == nil) {
		return fmt.Errorf("record continuation requires bounded explicit control stages")
	}
	for i, stage := range history {
		if i > 0 && stage.Control == nil {
			return fmt.Errorf("continued stage requires a Gooo policy")
		}
		rechecked, err := transitionRecordDependencies(ctx, plans, i, r)
		if err != nil {
			return err
		}
		if err := runRecordControlStage(ctx, plans[i], r, stage.Control); err != nil {
			return err
		}
		if rechecked != stage.RecheckedAttempts || len(r.Attempts) != stage.Attempts || r.SelectedMask != stage.SelectedMask || !reflect.DeepEqual(r.Control, stage.Control) {
			return fmt.Errorf("record control stage %d does not replay", i)
		}
		stage.Control = r.Control
		r.ControlHistory = append(r.ControlHistory, stage)
	}
	retained := len(r.Attempts)
	rechecked, err := transitionRecordDependencies(ctx, plans, len(history), r)
	if err != nil {
		return err
	}
	if err := runRecordControlStage(ctx, plans[len(history)], r, current); err != nil {
		return err
	}
	if len(history) > 0 {
		r.Continuation = &RecordAssemblyContinuation{RetainedAttempts: retained,
			RecheckedAttempts: rechecked, AddedAttempts: len(r.Attempts) - retained}
	}
	return nil
}

func runRecordControlStage(ctx context.Context, plan recordAssemblyPlan, r *RecordAssemblyReceipt, control *RecordAssemblyControl) error {
	var policy *preparedRecordPolicy
	r.Control = nil
	if control != nil {
		var err error
		policy, err = prepareRecordPolicy(ctx, control.Policy)
		if err != nil {
			return err
		}
		r.Control = policy.receipt
	}
	return searchRecordAssemblyWithPolicy(ctx, plan, r, policy)
}

// Reconstruct the best whole-case prefix separately from the previously
// selected candidate. Historical validation is not another candidate attempt.
func resumeRecordSearch(ctx context.Context, p recordAssemblyPlan, r *RecordAssemblyReceipt,
	policy *preparedRecordPolicy) (RecordAssemblyAttempt, []RecordAssemblyCase, []uint16, error) {
	remaining := r.Ranking[len(r.Attempts):]
	var best RecordAssemblyAttempt
	found := false
	for _, attempt := range r.Attempts {
		if attempt.Total > 0 && (!found || attempt.Passed > best.Passed ||
			(attempt.Passed == best.Passed && attempt.FieldsPassed > best.FieldsPassed)) {
			best, found = attempt, true
		}
	}
	var cases []RecordAssemblyCase
	if found {
		var err error
		_, cases, err = evaluateRecordMask(ctx, p, best.Mask)
		if err != nil {
			return best, nil, nil, err
		}
	}
	if r.Total > 0 && r.Passed == r.Total {
		remaining = nil
	}
	if policy != nil {
		entry, err := recordPolicyDecision(ctx, p, r, policy)
		if err != nil {
			return best, nil, nil, err
		}
		r.Control.Entry = &entry
		if found && entry.Operation == "USE_OBSERVED_CANDIDATE" {
			r.SelectedMask, r.Cases = best.Mask, cases
			r.Passed, r.Total, r.FieldsPassed, r.FieldsTotal = best.Passed, best.Total, best.FieldsPassed, best.FieldsTotal
		}
		if !entry.Continue {
			remaining = nil
		}
	}
	return best, cases, remaining, nil
}
