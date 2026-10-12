package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type criteriaQuestionEvaluator struct {
	plan       recordAssemblyPlan
	receipt    *CriteriaQuestionReceipt
	candidates map[uint16]generatedRoute
	outputs    map[string]int
	criteria   map[string]int
	predicates []*criteriaTupleEvaluator
}

func newCriteriaQuestionEvaluator(ctx context.Context, p recordAssemblyPlan, r *CriteriaQuestionReceipt) (*criteriaQuestionEvaluator, error) {
	e := &criteriaQuestionEvaluator{plan: p, receipt: r, candidates: map[uint16]generatedRoute{}, outputs: map[string]int{}, criteria: map[string]int{}}
	for _, predicate := range p.criteria {
		program := predicate.program
		prepared, err := newCriteriaTupleEvaluator(ctx, []byte(program.Source), program.Report.Activity, program.Report.RecordTypes, nil)
		if err != nil {
			return nil, err
		}
		e.predicates = append(e.predicates, prepared)
	}
	return e, nil
}

func criteriaQuestionKey(mask uint16, input string) (string, error) {
	key, err := assemblyspec.RecordValueKey(input)
	return fmt.Sprintf("%d:%s", mask, key), err
}

// Probe outputs carry no predicate verdict. Only acquired inputs become new
// obligations, so looking for disagreement does not pre-score unchosen inputs.
func (e *criteriaQuestionEvaluator) output(ctx context.Context, mask uint16, input string) (int, CriteriaQuestionOutput, error) {
	if err := ctx.Err(); err != nil {
		return 0, CriteriaQuestionOutput{}, err
	}
	key, err := criteriaQuestionKey(mask, input)
	if err != nil {
		return 0, CriteriaQuestionOutput{}, err
	}
	if index, ok := e.outputs[key]; ok && e.receipt.Options.ReuseProbeOutputs {
		e.receipt.ReusedOutputs++
		return index, e.receipt.Outputs[index], nil
	}
	row := CriteriaQuestionOutput{Mask: mask, Input: input}
	candidate, ok := e.candidates[mask]
	if !ok {
		candidate, err = e.plan.candidate(mask)
		if err == nil {
			e.candidates[mask] = candidate
		}
	}
	if err == nil {
		calls := newLocalCallRecorder(candidate.source, candidate.report, []string{input})
		calls.observation.Phase, calls.observation.CandidateID = "criteria-question-output", strconv.Itoa(int(mask))
		var evaluator *criteriaTupleEvaluator
		evaluator, err = newCriteriaTupleEvaluator(ctx, candidate.source, candidate.report.Activity, candidate.report.RecordTypes, calls)
		if err == nil {
			e.receipt.CandidateEvaluations++
			var raw json.RawMessage
			raw, err = evaluator.evaluate(input, 0)
			if err == nil {
				row.Value, err = assemblyspec.RecordValueKey(string(raw))
			}
		}
		row.LocalCalls = calls.finish(ctx, err)
	}
	if err != nil {
		row.Error = err.Error()
	}
	index := len(e.receipt.Outputs)
	e.receipt.Outputs = append(e.receipt.Outputs, row)
	if e.receipt.Options.ReuseProbeOutputs {
		e.outputs[key] = index
	}
	return index, row, ctx.Err()
}

func (e *criteriaQuestionEvaluator) evaluate(ctx context.Context, mask uint16, input string) (int, RecordCriteriaInput, error) {
	key, err := criteriaQuestionKey(mask, input)
	if err != nil {
		return 0, RecordCriteriaInput{}, err
	}
	if index, ok := e.criteria[key]; ok && e.receipt.Options.ReuseProbeOutputs {
		e.receipt.ReusedCriteriaResults++
		return index, e.receipt.Evaluations[index].Result, ctx.Err()
	}
	outputIndex, output, err := e.output(ctx, mask, input)
	if err != nil {
		return 0, RecordCriteriaInput{}, err
	}
	row := RecordCriteriaInput{InputJSON: input, ActualJSON: output.Value, Error: output.Error}
	var fault error
	if output.Error != "" {
		fault = fmt.Errorf("%s", output.Error)
	}
	for i, predicate := range e.plan.criteria {
		check := evaluateRecordCriterion(ctx, predicate, e.predicates[i], input, json.RawMessage(output.Value), 0, "question", fault)
		row.Checks = append(row.Checks, check)
		if check.LocalCalls != nil {
			e.receipt.PredicateEvaluations += check.LocalCalls.CasesStarted
		}
	}
	index := len(e.receipt.Evaluations)
	e.receipt.Evaluations = append(e.receipt.Evaluations, CriteriaQuestionEvaluation{Mask: mask, OutputIndex: outputIndex, Result: row})
	if e.receipt.Options.ReuseProbeOutputs {
		e.criteria[key] = index
	}
	return index, row, ctx.Err()
}
