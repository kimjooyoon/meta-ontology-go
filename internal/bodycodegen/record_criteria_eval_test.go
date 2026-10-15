package bodycodegen

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCriteriaValuesNeedNoInventedExpectedRecord(t *testing.T) {
	ctx := context.Background()
	source := []byte(typedCriteriaSource)
	owner, err := prepareActivityBody("criteria.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	base, err := newCriteriaTupleEvaluator(ctx, owner.base.source, "Select", owner.records, nil)
	if err != nil {
		t.Fatal(err)
	}
	value, err := base.evaluate("[9007199254740993]", 0)
	if err != nil || string(value) != `{"value":9007199254740993}` {
		t.Fatal("unscored record lost its exact value", string(value), err)
	}
	spec, _ := SourceAssembly(ctx, "criteria.gooo", source, "Select")
	prepared, err := prepareRecordCriteria(ctx, "criteria.gooo", source, owner, spec.Criteria)
	if err != nil {
		t.Fatal(err)
	}
	program := prepared[0].program
	predicate, err := newCriteriaTupleEvaluator(ctx, []byte(program.Source), program.Report.Activity, program.Report.RecordTypes, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ input, result string }{
		{`[9007199254740993,{"value":9007199254740993}]`, "false"},
		{`[9007199254740993,{"value":-9007199254740993}]`, "true"},
	} {
		value, err := predicate.evaluate(tc.input, 0)
		if err != nil || string(value) != tc.result {
			t.Fatal("predicate returned another relation", string(value), err)
		}
	}
	if _, err := predicate.evaluate(`[7,{"value":7,"unknown":true}]`, 0); err == nil {
		t.Fatal("predicate accepted undeclared record fields")
	}
}

func TestCriteriaEvaluationKeepsFaultsAndCancellationAsUnknownValues(t *testing.T) {
	source := []byte(strings.Replace(typedCriteriaSource, "input1.value == 0-input0", "input1.value / input0 > 0", 1))
	ctx, cancel := context.WithCancel(context.Background())
	program, err := GenerateWithPlanner(ctx, "criteria.gooo", source, "Negative", "", "")
	if err != nil {
		t.Fatal(err)
	}
	e, err := newCriteriaTupleEvaluator(ctx, []byte(program.Source), program.Report.Activity, program.Report.RecordTypes, nil)
	if err != nil {
		t.Fatal(err)
	}
	value, err := e.evaluate(`[0,{"value":7}]`, 0)
	if err == nil || value != nil {
		t.Fatal("predicate fault was converted to a Boolean", string(value), err)
	}
	value, err = e.evaluate(`[1,{"value":7}]`, 1)
	if err != nil || string(value) != "true" {
		t.Fatal("a previous fault changed the next input", string(value), err)
	}
	cancel()
	if value, err := e.evaluate(`[1,{"value":7}]`, 2); value != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("predicate cancellation was scored", string(value), err)
	}
}
