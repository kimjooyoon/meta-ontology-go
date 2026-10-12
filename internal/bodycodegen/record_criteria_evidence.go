package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
)

// Criteria counts are predicate/input pairs, independently of exact-value cases
// and individual record fields. Input JSON stays canonical and lossless.
type RecordCriteriaEvidence struct {
	Activity        string                   `json:"activity"`
	ActivityID      string                   `json:"activity_id"`
	Schema          string                   `json:"schema"`
	Phase           string                   `json:"phase"`
	SourceSHA256    string                   `json:"source_sha256"`
	GeneratedSHA256 string                   `json:"generated_sha256"`
	SuiteSHA256     string                   `json:"suite_sha256"`
	Bindings        []RecordCriterionBinding `json:"bindings"`
	Inputs          []RecordCriteriaInput    `json:"inputs"`
	Passed          int                      `json:"passed"`
	Failed          int                      `json:"failed"`
	Unknown         int                      `json:"unknown"`
	Total           int                      `json:"total"`
	InputsCompleted int                      `json:"inputs_completed"`
	InputsPassed    int                      `json:"inputs_passed"`
	InputsTotal     int                      `json:"inputs_total"`
	LocalCalls      *LocalCallObservation    `json:"local_calls"`
	Scope           string                   `json:"scope"`
}

type RecordCriterionBinding struct {
	ID              string `json:"id"`
	Activity        string `json:"activity"`
	ActivityID      string `json:"activity_id"`
	GeneratedSHA256 string `json:"generated_sha256"`
}

type RecordCriteriaInput struct {
	InputJSON  string                 `json:"input_json"`
	ActualJSON string                 `json:"actual_json,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Checks     []RecordCriterionCheck `json:"checks"`
}

type RecordCriterionCheck struct {
	ID         string                `json:"id"`
	Status     string                `json:"status"`
	Reason     string                `json:"reason,omitempty"`
	LocalCalls *LocalCallObservation `json:"local_calls,omitempty"`
}

func validateRecordCriteriaInputs(ctx context.Context, p recordAssemblyPlan) error {
	if p.spec.Criteria == nil {
		return nil
	}
	e, err := newCriteriaTupleEvaluator(ctx, p.body.base.source, p.body.activity.Name, p.body.records, nil)
	if err != nil {
		return err
	}
	for _, group := range [][]string{p.spec.Criteria.Inputs, p.spec.Criteria.HoldoutInputs} {
		for _, input := range group {
			if err := e.bind(input); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}

func evaluateRecordCriteria(ctx context.Context, p recordAssemblyPlan, candidate generatedRoute,
	inputs []string, phase string) (*RecordCriteriaEvidence, error) {
	if p.spec.Criteria == nil || len(inputs) == 0 {
		return nil, nil
	}
	raw, _ := json.Marshal(struct {
		Inputs   []string
		Criteria any
	}{inputs, p.spec.Criteria.Predicates})
	r := &RecordCriteriaEvidence{Schema: "gooo/record-criteria/v1", Phase: phase, Activity: candidate.report.Activity, ActivityID: candidate.report.ActivityID,
		SourceSHA256: p.sourceSHA256, GeneratedSHA256: digest(candidate.source), SuiteSHA256: digest(raw),
		InputsTotal: len(inputs), Total: len(inputs) * len(p.criteria), Unknown: len(inputs) * len(p.criteria),
		Scope: "declared predicates on finite input tuples; false and unavailable results remain separate; no natural-language intent coverage or training-independence claim"}
	for _, predicate := range p.criteria {
		r.Bindings = append(r.Bindings, RecordCriterionBinding{predicate.declaration.ID, predicate.declaration.Activity,
			predicate.activityID, predicate.goSHA256})
	}
	calls := newLocalCallRecorder(candidate.source, candidate.report, inputs)
	calls.observation.Phase = "criteria-" + phase
	e, err := newCriteriaTupleEvaluator(ctx, candidate.source, candidate.report.Activity, candidate.report.RecordTypes, calls)
	if err != nil {
		return nil, err
	}
	evaluators, err := prepareCriteriaEvaluators(ctx, p.criteria)
	if err != nil {
		return nil, err
	}
	var firstError error
	for i, input := range inputs {
		actual, fault := e.evaluate(input, i)
		row := RecordCriteriaInput{InputJSON: input, ActualJSON: string(actual)}
		if fault != nil {
			row.Error = fault.Error()
			if firstError == nil {
				firstError = fault
			}
		} else {
			r.InputsCompleted++
		}
		allPassed := true
		for j, predicate := range p.criteria {
			check := evaluateRecordCriterion(ctx, predicate, evaluators[j], input, actual, i, phase, fault)
			row.Checks = append(row.Checks, check)
			switch check.Status {
			case "SATISFIED":
				r.Passed++
				r.Unknown--
			case "UNSATISFIED":
				r.Failed++
				r.Unknown--
				allPassed = false
			default:
				allPassed = false
			}
		}
		if allPassed {
			r.InputsPassed++
		}
		r.Inputs = append(r.Inputs, row)
		if err := ctx.Err(); err != nil {
			r.LocalCalls = calls.finish(ctx, err)
			return r, err
		}
	}
	r.LocalCalls = calls.finish(ctx, firstError)
	return r, ctx.Err()
}

func evaluateRecordCriterion(ctx context.Context, predicate preparedRecordCriterion, e *criteriaTupleEvaluator, input string,
	actual json.RawMessage, caseIndex int, phase string, fault error) RecordCriterionCheck {
	check := RecordCriterionCheck{ID: predicate.declaration.ID, Status: "UNKNOWN"}
	if fault != nil {
		check.Reason = "CANDIDATE_UNAVAILABLE: " + fault.Error()
		return check
	}
	var values []json.RawMessage
	if err := json.Unmarshal([]byte(input), &values); err != nil {
		check.Reason = err.Error()
		return check
	}
	values = append(values, actual)
	tuple, err := json.Marshal(values)
	if err != nil {
		check.Reason = err.Error()
		return check
	}
	program := predicate.program
	calls := newLocalCallRecorder([]byte(program.Source), program.Report, []string{string(tuple)})
	calls.observation.Phase = "criterion-" + phase
	e.evaluator.localCalls = calls
	value, err := e.evaluate(string(tuple), caseIndex)
	check.LocalCalls = calls.finish(ctx, err)
	if err != nil {
		check.Reason = "PREDICATE_UNAVAILABLE: " + err.Error()
		return check
	}
	switch string(value) {
	case "true":
		check.Status = "SATISFIED"
	case "false":
		check.Status = "UNSATISFIED"
	default:
		check.Reason = fmt.Sprintf("predicate returned non-Boolean %s", value)
	}
	return check
}
