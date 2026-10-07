package bodycodegen

import (
	"encoding/json"
	"fmt"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func recordOrdinalContext(choices []RecordValueChoice) *RecordOrdinalContext {
	r := &RecordOrdinalContext{Schema: "gooo/record-field-ordinal-context/v1", Status: "ENCODED",
		Scope: "typed source-bound field expressions mapped to ordinal binary choices; frozen integer model transfer; cases and expected outputs excluded"}
	for _, choice := range choices {
		raw, _ := json.Marshal(struct {
			Field    string `json:"field"`
			TypeID   string `json:"type_id,omitempty"`
			Presence string `json:"presence,omitempty"`
			First    string `json:"first"`
			Second   string `json:"second"`
			Intent   string `json:"intent"`
		}{choice.FieldID, choice.TypeID, choice.Presence, choice.First, choice.Second, choice.Intent})
		r.Parts = append(r.Parts, string(raw))
	}
	if len(choices) != 3 {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_FIELD_CHOICES_REQUIRED"
		return r
	}
	fields, err := ordinalProxyFeatures()
	if err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", err.Error()
		return r
	}
	var parts [3]string
	for i, part := range r.Parts {
		parts[i], err = decision.EncodeSemanticContextInput(fields, part)
		if err != nil {
			r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "COMPLETE_FIELD_CONTEXT_EXCEEDS_MODEL_BOUND"
			return r
		}
	}
	r.Text, err = jointdecision.EncodeThree(parts)
	if err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", err.Error()
		return r
	}
	r.SHA256 = digest([]byte(r.Text))
	return r
}

// This pure integer proxy represents the ordinal alternatives 0 and 1. Its
// source feature identity is distinct from actual record semantics, carried in
// the complete natural text and the source-bound record choice receipt.
func ordinalProxyFeatures() ([decision.SplitContextDim]byte, error) {
	plan := pathplan.Plan{Schema: pathplan.Schema, Base: bodyplan.Plan{Schema: bodyplan.Schema, ID: "gooo://record-field-ordinal/v1",
		Name: "FieldOrdinal", ResultType: decision.TypeInt,
		Expressions: []bodyplan.Expr{{Kind: bodyplan.ExprInput, Name: "input"}, {Kind: bodyplan.ExprInt, Int: 0}, {Kind: bodyplan.ExprInt, Int: 1},
			{Kind: bodyplan.ExprLocal, Name: "first"}, {Kind: bodyplan.ExprLocal, Name: "second"},
			{Kind: bodyplan.ExprBinary, Operation: "add", Left: 3, Right: 4},
			{Kind: bodyplan.ExprLocal, Name: "first"}, {Kind: bodyplan.ExprBinary, Operation: "add", Left: 5, Right: 6}},
		Statements: []bodyplan.Stmt{{Kind: bodyplan.StmtLet, Name: "first", Expr: 1}, {Kind: bodyplan.StmtLet, Name: "second", Expr: 2},
			{Kind: bodyplan.StmtReturn, Expr: 7}}, Root: []int{0, 1, 2}},
		Decisions: []pathplan.Choice{{ID: "field-ordinal", Kind: pathplan.LocalReference, Target: 3, Intent: "Choose a permitted field value ordinal.",
			Fallback: "reference_first", Options: []pathplan.Option{{Label: "reference_first", Name: "first"}, {Label: "reference_second", Name: "second"}}}}}
	prepared, err := pathplan.Prepare(plan)
	if err != nil {
		return [decision.SplitContextDim]byte{}, fmt.Errorf("ordinal proxy: %w", err)
	}
	return prepared.SourceFeatures("field-ordinal")
}
