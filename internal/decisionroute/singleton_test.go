package decisionroute

import "testing"

func TestSingletonValidationKeepsOrdinaryChoiceBoundary(t *testing.T) {
	r := Request{Schema: RequestSchema, State: "{}", Fallback: "only", Question: Question{
		ID: "fill", Instructions: "Use the only checked assignment.", Options: []Option{{ID: "only", Description: "A complete assignment."}},
	}}
	if _, err := Validate(r); err == nil {
		t.Fatal("ordinary provider request accepted a singleton")
	}
	digest, err := ValidateSingleton(r)
	if err != nil || digest == "" {
		t.Fatal(digest, err)
	}
	r.State = `{"rejected":"bad"}`
	other, err := ValidateSingleton(r)
	if err != nil || digest == other {
		t.Fatal("request digest omitted its rejection state")
	}
	r.Fallback = "missing"
	if _, err := ValidateSingleton(r); err == nil {
		t.Fatal("undeclared fallback accepted")
	}
	r.Fallback = "only"
	r.Question.Options = append(r.Question.Options, Option{ID: "other", Description: "Other assignment."})
	if _, err := ValidateSingleton(r); err == nil {
		t.Fatal("multiple singleton options accepted")
	}
}
