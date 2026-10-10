package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const pathConditionSchema = "gooo/typed-path-conditions/v1"

type PathConditionReceipt struct {
	Schema             string                     `json:"schema"`
	SuiteSHA256        string                     `json:"suite_sha256"`
	Declared           int                        `json:"declared"`
	Passed             int                        `json:"passed"`
	NotReached         int                        `json:"not_reached"`
	Status             string                     `json:"status"`
	Results            []pathplan.ConditionResult `json:"results,omitempty"`
	SelectedTreeSHA256 string                     `json:"selected_tree_sha256,omitempty"`
	EmittedTreeSHA256  string                     `json:"emitted_tree_sha256,omitempty"`
}

func sameConditionResults(left, right []pathplan.ConditionResult) bool {
	return reflect.DeepEqual(left, right)
}

func inspectPathConditions(ctx context.Context, prepared *pathplan.PreparedPlan,
	choices map[string]string) (*PathConditionReceipt, error) {
	if !prepared.HasConditions() {
		return nil, nil
	}
	rows, err := prepared.CheckDeclaredConditions(ctx, choices)
	if err != nil {
		return nil, err
	}
	cases := make([]pathplan.ConditionCase, len(rows))
	receipt := &PathConditionReceipt{Schema: pathConditionSchema, Declared: len(rows), Results: rows, Status: "MATCHED"}
	for i, row := range rows {
		cases[i] = row.Case
		if row.Passed {
			receipt.Passed++
		}
		if !row.Observation.Reached {
			receipt.NotReached++
		}
	}
	raw, _ := json.Marshal(cases)
	receipt.SuiteSHA256 = digest(raw)
	if !pathplan.ConditionsPassed(rows) {
		receipt.Status = "CONDITION_REJECTED"
		return receipt, fmt.Errorf("declared conditions matched %d/%d; %d not reached", receipt.Passed, receipt.Declared, receipt.NotReached)
	}
	return receipt, nil
}

// Structural equality checks every emitted operator and ordered child edge,
// including intermediate predicates whose changes can cancel at the output.
// The SDK observations apply to these finite inputs; this is not an NL oracle.
func selectedPathConditions(ctx context.Context, prepared *pathplan.PreparedPlan, choices map[string]string,
	selected *bodyplan.Program, activity string, generated []byte) (*PathConditionReceipt, error) {
	receipt, err := inspectPathConditions(ctx, prepared, choices)
	if err != nil || receipt == nil {
		return receipt, err
	}
	left, err := typedBodyTree(ctx, activity, []byte(selected.GoSource()))
	if err != nil {
		return receipt, err
	}
	right, err := typedBodyTree(ctx, activity, generated)
	if err != nil {
		return receipt, err
	}
	receipt.SelectedTreeSHA256, receipt.EmittedTreeSHA256 = digest(left), digest(right)
	if !bytes.Equal(left, right) {
		receipt.Status = "STRUCTURE_MISMATCH"
		return receipt, fmt.Errorf("emitted body changes the selected condition or program structure")
	}
	return receipt, nil
}

func pathConditionDimension(p *BodyPathReceipt) CompletenessDimension {
	c := p.Conditions
	cases := make([]pathplan.ConditionCase, len(c.Results))
	passed, unreached := 0, 0
	mismatch := c.Schema != pathConditionSchema || c.Declared < 1 || c.Declared > 128 || !validDigest(c.SuiteSHA256)
	for i, row := range c.Results {
		cases[i] = row.Case
		matched := row.Observation.Reached && row.Observation.Value == row.Case.Expected
		passed += boolCount(matched)
		unreached += boolCount(!row.Observation.Reached)
		status := "MISMATCH"
		if !row.Observation.Reached {
			status = "NOT_REACHED"
		} else if matched {
			status = "MATCH"
		}
		mismatch = mismatch || row.Passed != matched || row.Status != status
	}
	if len(c.Results) > 0 {
		raw, _ := json.Marshal(cases)
		mismatch = mismatch || digest(raw) != c.SuiteSHA256 || len(c.Results) != c.Declared ||
			passed != c.Passed || unreached != c.NotReached || c.Status != "MATCHED" ||
			!validDigest(c.SelectedTreeSHA256) || c.SelectedTreeSHA256 != c.EmittedTreeSHA256
	}
	return completenessDimension("typed_path_intermediate_conditions", passed, c.Declared,
		"declared finite if-condition expectations matched by the selected program",
		"Skipped conditions remain unobserved; selected and emitted typed trees must agree. Counts cover only the source-declared predicate cases.",
		[]string{"body_paths.conditions", "condition_suite_sha256:" + c.SuiteSHA256}, mismatch)
}
