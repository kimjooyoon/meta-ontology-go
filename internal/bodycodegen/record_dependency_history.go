package bodycodegen

import (
	"context"
	"errors"
	"fmt"
	"go/types"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Checkpoint changes may alter fixed helper bodies, never the target's choices,
// obligations, signature or source-owned budget. Ignore syntax offsets.
func compatibleRecordTarget(a, b recordAssemblyPlan) error {
	x, _ := a.spec.Canonical()
	y, _ := b.spec.Canonical()
	if x != y || a.body.body != b.body.body || a.body.packageName != b.body.packageName ||
		a.body.activityID != b.body.activityID || a.body.outputType != b.body.outputType ||
		!reflect.DeepEqual(a.body.parameters, b.body.parameters) ||
		!reflect.DeepEqual(a.body.allRecords, b.body.allRecords) ||
		!reflect.DeepEqual(a.body.activityIDs, b.body.activityIDs) ||
		!reflect.DeepEqual(recordActivitySignatures(a), recordActivitySignatures(b)) ||
		!reflect.DeepEqual(a.choices, b.choices) {
		return fmt.Errorf("resume target contract or typed declarations changed")
	}
	return nil
}

func recordActivitySignatures(p recordAssemblyPlan) map[string][]string {
	result := make(map[string][]string)
	for _, declaration := range p.body.file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok {
			signature := []string{activity.Output}
			for _, input := range activity.Inputs {
				signature = append(signature, input.Name)
			}
			result[activity.Name] = signature
		}
	}
	return result
}

func appendRecordHistory(old *RecordAssemblyReceipt, priorSource, source []byte) []RecordAssemblyControlStage {
	history := append([]RecordAssemblyControlStage(nil), old.ControlHistory...)
	stage := RecordAssemblyControlStage{Attempts: len(old.Attempts), SelectedMask: old.SelectedMask, Control: old.Control}
	if old.Continuation != nil {
		stage.RecheckedAttempts = old.Continuation.RecheckedAttempts
	}
	// Omitted checkpoints inherit the previous source. Legacy histories all used
	// priorSource; materialize the first one when that source starts changing.
	if len(history) > 0 && history[0].Source == "" && string(priorSource) != string(source) {
		history[0].Source, history[0].SourceSHA256 = string(priorSource), digest(priorSource)
	}
	previous := string(source)
	for _, saved := range history {
		if saved.Source != "" {
			previous = saved.Source
		}
	}
	if previous != string(priorSource) {
		stage.Source, stage.SourceSHA256 = string(priorSource), digest(priorSource)
	}
	return append(history, stage)
}

func recordHistoryPlans(ctx context.Context, filename string, source []byte, current recordAssemblyPlan,
	history []RecordAssemblyControlStage) ([]recordAssemblyPlan, string, error) {
	if len(source) > 128<<10 || len(history) > recordControlStageLimit {
		return nil, "", fmt.Errorf("record continuation source or history exceeds its bound")
	}
	plan := current
	plans := make([]recordAssemblyPlan, 0, len(history)+1)
	origin := ""
	for i, stage := range history {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}
		if stage.Source != "" {
			if len(stage.Source) > 128<<10 || stage.SourceSHA256 != digest([]byte(stage.Source)) ||
				(i > 0 && history[0].Source == "") {
				return nil, "", fmt.Errorf("record source checkpoint %d is unbound", i)
			}
			var err error
			plan, err = prepareRecordAssembly(ctx, filename, []byte(stage.Source), current.body.activity.Name)
			if err != nil {
				return nil, "", err
			}
			if err := compatibleRecordTarget(plan, current); err != nil {
				return nil, "", err
			}
			if i == 0 {
				origin = stage.SourceSHA256
			}
		} else if stage.SourceSHA256 != "" {
			return nil, "", fmt.Errorf("record checkpoint digest has no source")
		}
		plans = append(plans, plan)
	}
	return append(plans, current), origin, nil
}

func evaluateRecordMask(ctx context.Context, p recordAssemblyPlan, mask uint16) (RecordAssemblyAttempt, []RecordAssemblyCase, error) {
	if err := ctx.Err(); err != nil {
		return RecordAssemblyAttempt{}, nil, err
	}
	candidate, err := p.candidate(mask)
	if err != nil {
		if contextError := ctx.Err(); contextError != nil {
			return RecordAssemblyAttempt{}, nil, contextError
		}
		if _, ok := errors.AsType[types.Error](err); ok {
			return RecordAssemblyAttempt{Mask: mask, Status: "TYPECHECK_FAILED", Reason: err.Error()}, nil, nil
		}
		return RecordAssemblyAttempt{}, nil, err
	}
	cases, err := evaluateRecordAssembly(ctx, candidate.source, p.body.activity.Name, p.body.records, p.spec.ValueCases)
	return scoreRecordCases(mask, cases), cases, err
}

func transitionRecordDependencies(ctx context.Context, plans []recordAssemblyPlan, index int, r *RecordAssemblyReceipt) (int, error) {
	if index == 0 || plans[index-1].dependencySHA256 == plans[index].dependencySHA256 {
		return 0, nil
	}
	// Evaluate each retained mask under the new closure, even if an early one
	// passes. These are rechecks of attempted identities, not new budget slots.
	r.Cases = nil
	r.SelectedMask, r.Passed, r.Total, r.FieldsPassed, r.FieldsTotal = 0, 0, 0, 0, 0
	for i, prior := range r.Attempts {
		attempt, cases, err := evaluateRecordMask(ctx, plans[index], prior.Mask)
		if err != nil {
			return 0, err
		}
		r.Attempts[i] = attempt
		if attempt.Total > 0 && (r.Cases == nil || attempt.FieldsPassed > r.FieldsPassed ||
			(attempt.FieldsPassed == r.FieldsPassed && attempt.Passed > r.Passed)) {
			r.SelectedMask, r.Cases = attempt.Mask, cases
			r.Passed, r.Total, r.FieldsPassed, r.FieldsTotal = attempt.Passed, attempt.Total, attempt.FieldsPassed, attempt.FieldsTotal
		}
	}
	return len(r.Attempts), nil
}
