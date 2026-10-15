package bodycodegen

import (
	"context"
	"fmt"
	"strings"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Candidate compilation is separate from the SDK's exact-case search document.
// The Gooo predicates own the obligations, including a criteria-only program.
type scalarCriteriaCandidateError struct{ error }

func (e *scalarCriteriaCandidateError) Unwrap() error { return e.error }

type scalarCriteriaPlan struct {
	sourcePlan pathplan.Plan
	prepared   *pathplan.PreparedPlan
	choices    []pathplan.Choice
}

func prepareScalarCriteriaAssembly(ctx context.Context, filename string, source []byte,
	activity string, p recordAssemblyPlan) (recordAssemblyPlan, error) {
	if p.body.outputType != "int64" || len(p.body.parameters) != 1 || p.body.parameters[0].Type != "int64" {
		return p, fmt.Errorf("scalar path criteria require Integer -> Integer")
	}
	base := Result{Source: string(p.body.base.source), Report: p.body.base.report}
	document, err := sourceRecipeDocument(ctx, base, activity, sourceAssemblyRecipe(p.spec))
	if err != nil {
		return p, err
	}
	prepared, err := pathplan.Prepare(document.Plan)
	if err != nil {
		return p, err
	}
	if _, err = bindTypedPathSource(ctx, filename, source, activity, prepared, &BodyPathReceipt{}); err != nil {
		return p, fmt.Errorf("scalar criteria source binding: %w", err)
	}
	p.scalar = &scalarCriteriaPlan{prepared: prepared, choices: document.Plan.Decisions, sourcePlan: document.Plan}
	for i, choice := range document.Plan.Decisions {
		p.choices = append(p.choices, RecordValueChoice{ID: choice.ID, Kind: choice.Kind,
			Occurrence: p.spec.Choices[i].Occurrence, Intent: choice.Intent,
			First: choice.Options[0].Label, Second: choice.Options[1].Label})
	}
	p.criteria, err = prepareRecordCriteria(ctx, filename, source, p.body, p.spec.Criteria)
	if err != nil {
		return p, err
	}
	if err = validateRecordCriteriaInputs(ctx, p); err != nil {
		return p, err
	}
	for _, held := range []bool{false, true} {
		if err = validateRecordAssemblyCases(ctx, p.body.base.source, activity, p.body.records, criteriaValueCases(p.spec, held)); err != nil {
			return p, err
		}
	}
	if err = validateRecordObservationProbes(ctx, p); err != nil {
		return p, err
	}
	var dependencies strings.Builder
	dependencies.WriteString(prepared.PlanSHA256())
	for _, predicate := range p.criteria {
		dependencies.WriteString(predicate.activityID)
		dependencies.WriteString(predicate.goSHA256)
	}
	p.dependencySHA256 = digest([]byte(dependencies.String()))
	return p, ctx.Err()
}

func (p *scalarCriteriaPlan) body(mask uint16) (string, error) {
	if int(mask) >= 1<<len(p.choices) {
		return "", fmt.Errorf("scalar choice mask exceeds declared alternatives")
	}
	selected := make(map[string]string, len(p.choices))
	for i, choice := range p.choices {
		index := 0
		if mask&(1<<i) != 0 {
			index = 1
		}
		selected[choice.ID] = choice.Options[index].Label
	}
	program, err := p.prepared.Compile(selected)
	if err != nil {
		return "", &scalarCriteriaCandidateError{err}
	}
	return program.GoooBody(), nil
}

func assemblyChoiceLabels(kind string) (string, string) {
	switch kind {
	case "operand_order", "branch_layout":
		return "layout_forward", "layout_reverse"
	case "local_reference":
		return "reference_first", "reference_second"
	case "assignment_target":
		return "assign_first", "assign_second"
	case "root_order":
		return "schedule_forward", "schedule_reverse"
	default:
		return "value_first", "value_second"
	}
}
