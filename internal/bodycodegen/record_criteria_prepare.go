package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type preparedRecordCriterion struct {
	declaration assemblyspec.Criterion
	activityID  string
	goSHA256    string
	program     Result
}

// Prepare fixed predicate programs before ranking or executing candidates.
// Positional types are compared independently of input/input0 spelling.
func prepareRecordCriteria(ctx context.Context, filename string, source []byte, owner preparedBody,
	criteria *assemblyspec.Criteria) ([]preparedRecordCriterion, error) {
	if ctx == nil {
		return nil, fmt.Errorf("criteria requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if criteria == nil {
		return nil, nil
	}
	if len(owner.parameters) >= 16 {
		return nil, fmt.Errorf("criteria needs one result parameter after at most 15 assembly inputs")
	}
	result := make([]preparedRecordCriterion, 0, len(criteria.Predicates))
	for _, predicate := range criteria.Predicates {
		prepared, err := prepareRecordCriterion(ctx, filename, source, owner, predicate)
		if err != nil {
			return nil, fmt.Errorf("criterion %q: %w", predicate.ID, err)
		}
		result = append(result, prepared)
	}
	return result, ctx.Err()
}

func prepareRecordCriterion(ctx context.Context, filename string, source []byte, owner preparedBody,
	predicate assemblyspec.Criterion) (preparedRecordCriterion, error) {
	if predicate.Activity == owner.activity.Name {
		return preparedRecordCriterion{}, fmt.Errorf("predicate must be a separate fixed computes activity")
	}
	spec, err := SourceAssembly(ctx, filename, source, predicate.Activity)
	if err != nil {
		return preparedRecordCriterion{}, err
	}
	if spec != nil {
		return preparedRecordCriterion{}, fmt.Errorf("predicate requires fixed computes")
	}
	declared, err := prepareActivityBodySource(filename, source, predicate.Activity)
	if err != nil {
		return preparedRecordCriterion{}, err
	}
	if declared.outputType != "bool" || len(declared.parameters) != len(owner.parameters)+1 {
		return preparedRecordCriterion{}, fmt.Errorf("predicate must take the assembly inputs followed by its result and return Boolean")
	}
	for i, input := range owner.parameters {
		if declared.parameters[i].Type != input.Type {
			return preparedRecordCriterion{}, fmt.Errorf("predicate input %d must match the assembly input type", i)
		}
	}
	if declared.parameters[len(owner.parameters)].Type != owner.outputType {
		return preparedRecordCriterion{}, fmt.Errorf("predicate last input must match the assembly result type")
	}
	program, err := GenerateWithPlanner(ctx, filename, source, predicate.Activity, "", "")
	if err != nil {
		return preparedRecordCriterion{}, fmt.Errorf("prepare fixed predicate closure: %w", err)
	}
	return preparedRecordCriterion{declaration: predicate, activityID: program.Report.ActivityID,
		goSHA256: digest([]byte(program.Source)), program: program}, nil
}
