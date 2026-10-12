package toolchainrelease

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateRelativeV2Candidates(p *bodycodegen.BodyPathReceipt, source []byte) error {
	d, err := bodycodegen.DecodeSourcePathDocument(context.Background(), "relative-v2.gooo", source, "Choose", nil)
	if err != nil {
		return err
	}
	if len(p.Search.Attempts) < 1 || len(p.Search.Attempts) > 4 {
		return fmt.Errorf("V2 local search history missing")
	}
	prepared, err := d.Prepare()
	if err != nil {
		return err
	}
	for _, attempt := range p.Search.Attempts {
		if err = validateRelativeV2Candidate(prepared, d, attempt); err != nil {
			return err
		}
	}
	final := p.Search.Attempts[len(p.Search.Attempts)-1]
	if final.Passed != 6 || final.Total != 6 || len(final.Results) != 6 || len(final.Conditions) != 3 {
		return fmt.Errorf("V2 final local denominator differs")
	}
	for i, x := range final.Results {
		q := d.TestCases[i]
		if x.Input != q.Input || x.Expected != q.Expected || x.Actual != q.Expected || !x.Passed {
			return fmt.Errorf("V2 local output%d differs", i)
		}
	}
	for i, x := range final.Conditions {
		if !reflect.DeepEqual(x.Case, d.Plan.ConditionCases[i]) || !x.Observation.Reached || x.Observation.Value != x.Case.Expected || !x.Passed {
			return fmt.Errorf("V2 condition%d differs", i)
		}
	}
	return nil
}

func validateRelativeV2Candidate(prepared *pathplan.PreparedPlan, d pathplan.Document, attempt pathplan.SearchAttempt) error {
	program, err := prepared.Compile(attempt.Choices)
	if err != nil {
		return err
	}
	if len(attempt.Results) != 6 || len(attempt.Conditions) != 3 || attempt.Total != 6 {
		return fmt.Errorf("V2 candidate history denominator differs")
	}
	passed := 0
	for i, row := range attempt.Results {
		q := d.TestCases[i]
		value, err := program.Evaluate(q.Input)
		if err != nil {
			return err
		}
		if row.Input != q.Input || row.Expected != q.Expected || row.Actual != value.Int || row.Passed != (value.Int == q.Expected) {
			return fmt.Errorf("V2 recorded candidate value differs")
		}
		if row.Passed {
			passed++
		}
	}
	conditions, err := prepared.CheckConditions(context.Background(), attempt.Choices, d.Plan.ConditionCases)
	if err != nil {
		return err
	}
	if attempt.Passed != passed || !reflect.DeepEqual(attempt.Conditions, conditions) {
		return fmt.Errorf("V2 candidate Boolean history differs")
	}
	status := "EVALUATED"
	for _, condition := range conditions {
		if !condition.Passed {
			status = "CONDITION_REJECTED"
		}
	}
	if attempt.Status != status {
		return fmt.Errorf("V2 recorded condition acceptance differs")
	}
	return nil
}
