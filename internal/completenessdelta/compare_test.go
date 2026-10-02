package completenessdelta

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

func fixture(state string, n, total int) completeness.CompletenessReceipt {
	d := completeness.CompletenessDimension{ID: "finite", Status: state, Numerator: n, Denominator: total,
		Unit: "finite expectations", Reason: "recorded observation", Evidence: []string{"fixture evidence"}}
	r := completeness.CompletenessReceipt{Schema: completeness.CompletenessReceiptSchema, ProfileID: generationProfile,
		Decision: "PROGRESS", DecisionBasis: "fixture observations", CoreDimensions: []string{d.ID},
		Dimensions: []completeness.CompletenessDimension{d}, NotClaimed: []string{},
		Scope: map[string]any{"receipt_declaration": completeness.ContractBinding(),
			"domain_scope": "fixture integer activity", "input_type": "int64", "output_type": "int64",
			"allowed_investment": "finite search", "system_budget": map[string]any{"human_actions": 0},
			"excluded_scope": []string{"unseen behavior"}, "boundary": map[string]any{"non_authorizing": true},
			"typed_path": map[string]any{"case_evaluator": "bounded_integer_go_ast", "search_config": map[string]any{"steps": 1},
				"original_source_sha256": digest([]byte("source")), "document_sha256": digest([]byte("document")),
				"test_suite_sha256": digest([]byte("cases")), "search_config_sha256": digest([]byte(`{"steps":1}`))}}}
	account(&r)
	return r
}

func account(r *completeness.CompletenessReceipt) {
	r.StatusCounts = map[string]int{"PASS": 0, "PROGRESS": 0, "UNKNOWN": 0, "FAIL_CLOSED": 0}
	r.UnresolvedClaims = []completeness.UnresolvedCompletenessClaim{}
	r.FirstUnresolved, r.FailClosedReason, r.Decision = nil, nil, "PROGRESS"
	for _, d := range r.Dimensions {
		r.StatusCounts[d.Status]++
		if d.Status != "PASS" {
			r.UnresolvedClaims = append(r.UnresolvedClaims, completeness.UnresolvedCompletenessClaim{ID: d.ID, Status: d.Status, Reason: d.Reason, NextOperation: "observe"})
		}
		if d.Status == "FAIL_CLOSED" {
			reason := "observed failure"
			r.FailClosedReason, r.Decision = &reason, "FAIL_CLOSED"
		}
	}
	if len(r.UnresolvedClaims) > 0 {
		c := r.UnresolvedClaims[0]
		r.FirstUnresolved = &c
	}
}

func encode(t *testing.T, r completeness.CompletenessReceipt) []byte {
	t.Helper()
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAllStateTransitionsPreserveUnknownAndFailure(t *testing.T) {
	states := []string{"PASS", "PROGRESS", "UNKNOWN", "FAIL_CLOSED"}
	for _, before := range states {
		for _, after := range states {
			t.Run(before+"-"+after, func(t *testing.T) {
				a, b := encode(t, fixture(before, 1, 1)), encode(t, fixture(after, 1, 1))
				aCopy, bCopy := bytes.Clone(a), bytes.Clone(b)
				r, err := Compare(a, b)
				if err != nil {
					t.Fatal(err)
				}
				d := r.Dimensions[0]
				wantNumeric := (before == "PASS" || before == "PROGRESS") && (after == "PASS" || after == "PROGRESS")
				if r.Relation != "SAME_MEASUREMENT_SCOPE" || (d.CountMagnitude != nil) != wantNumeric ||
					d.StateTransition != before+"_TO_"+after || !bytes.Equal(a, aCopy) || !bytes.Equal(b, bCopy) {
					t.Fatalf("transition changed meaning: %+v", d)
				}
				if before == "UNKNOWN" && (after == "PASS" || after == "PROGRESS") && d.ObservationChange != "BECAME_OBSERVED" {
					t.Fatal(d)
				}
				if (before == "PASS" || before == "PROGRESS") && after == "UNKNOWN" && d.Regression != "OBSERVATION_LOST" {
					t.Fatal(d)
				}
				if before == "FAIL_CLOSED" && after == "UNKNOWN" && (d.Regression != "FAILURE_EVIDENCE_LOST" || d.ObservationChange != "FAILURE_NOW_UNKNOWN") {
					t.Fatal("missing failure evidence was treated as recovery", d)
				}
				again, err := Compare(a, b)
				if err != nil || !reflect.DeepEqual(r, again) {
					t.Fatal("nondeterministic comparison", err)
				}
				for _, n := range r.ComparatorOperations {
					if n != 0 {
						t.Fatal("comparator performed external work")
					}
				}
			})
		}
	}
}

func TestExactCountsRegressionsAndScopeChanges(t *testing.T) {
	a := fixture("PROGRESS", math.MaxInt-2, math.MaxInt)
	b := fixture("PROGRESS", math.MaxInt-1, math.MaxInt)
	a.Scope["large_integer"], b.Scope["large_integer"] = json.Number("9007199254740993"), json.Number("9007199254740993")
	r, err := Compare(encode(t, a), encode(t, b))
	if err != nil {
		t.Fatal(err)
	}
	if *r.Dimensions[0].CountMagnitude != 1 || r.Dimensions[0].CountDirection != "INCREASED" || r.BeforeReceipt["scope"].(map[string]any)["large_integer"] != json.Number("9007199254740993") {
		t.Fatal("integer precision lost")
	}
	r, err = Compare(encode(t, b), encode(t, a))
	if err != nil || r.Dimensions[0].Regression != "COUNT_REGRESSION" {
		t.Fatal(r, err)
	}
	for name, mutate := range map[string]func(*completeness.CompletenessReceipt){
		"unit":        func(r *completeness.CompletenessReceipt) { r.Dimensions[0].Unit = "other" },
		"denominator": func(r *completeness.CompletenessReceipt) { r.Dimensions[0].Denominator-- },
		"suite": func(r *completeness.CompletenessReceipt) {
			r.Scope["typed_path"].(map[string]any)["test_suite_sha256"] = digest([]byte("changed"))
		},
		"missing-source": func(r *completeness.CompletenessReceipt) {
			delete(r.Scope["typed_path"].(map[string]any), "original_source_sha256")
		},
		"investment":          func(r *completeness.CompletenessReceipt) { r.Scope["allowed_investment"] = "larger budget" },
		"unsupported-profile": func(r *completeness.CompletenessReceipt) { r.ProfileID = "future/v1" },
		"unbound-digest": func(r *completeness.CompletenessReceipt) {
			r.Scope["typed_path"].(map[string]any)["document_sha256"] = "same"
		},
	} {
		t.Run(name, func(t *testing.T) {
			before, after := fixture("PROGRESS", 1, 4), fixture("PROGRESS", 2, 4)
			mutate(&after)
			r, err := Compare(encode(t, before), encode(t, after))
			if err != nil {
				t.Fatal(err)
			}
			if r.Dimensions[0].CountMagnitude != nil || r.Dimensions[0].CountDirection != "NOT_COMPARABLE" {
				t.Fatal("incomparable observation scored")
			}
		})
	}
}

func TestParentBytesNewObservationsAndRemovedObligations(t *testing.T) {
	before := fixture("UNKNOWN", 0, 1)
	raw := append(encode(t, before), '\n')
	after := fixture("PASS", 1, 1)
	after.ProfileID = runtimeProfile
	after.Scope["parent_receipt_sha256"] = digest(raw)
	after.Dimensions = append(after.Dimensions, completeness.CompletenessDimension{ID: "runtime", Status: "PASS", Numerator: 2, Denominator: 2, Unit: "runs", Reason: "completed", Evidence: []string{"output"}})
	account(&after)
	r, err := Compare(raw, encode(t, after))
	if err != nil {
		t.Fatal(err)
	}
	if r.Relation != "PARENT_RUNTIME_CONTINUATION" || r.Dimensions[0].ObservationChange != "BECAME_OBSERVED" || r.Dimensions[0].CountMagnitude != nil || r.Dimensions[1].Presence != "ADDED" {
		t.Fatal(r)
	}
	r, err = Compare(bytes.TrimSpace(raw), encode(t, after))
	if err != nil || r.Relation != "INCOMPARABLE_SCOPE" {
		t.Fatal("parent whitespace binding lost", err)
	}
	removed := fixture("PASS", 1, 1)
	removed.Dimensions[0].ID = "replacement"
	removed.CoreDimensions = []string{"replacement"}
	account(&removed)
	r, err = Compare(raw, encode(t, removed))
	if err != nil {
		t.Fatal(err)
	}
	if r.Dimensions[0].Presence != "REMOVED" || r.Dimensions[0].Regression != "OBLIGATION_REMOVED" || r.Dimensions[1].Presence != "ADDED" {
		t.Fatal(r)
	}
}

func TestStrictInputsAndGeneratedArtifacts(t *testing.T) {
	raw := encode(t, fixture("PASS", 1, 1))
	for _, bad := range [][]byte{
		append(bytes.Clone(raw), []byte("{}")...), bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1),
		bytes.Replace(raw, []byte(`"profile_id":`), []byte(`"Profile_ID":`), 1), {0xff}, []byte("null"),
		bytes.Repeat([]byte("x"), MaxInputBytes+1), []byte(strings.Repeat("[", 66) + "0" + strings.Repeat("]", 66)),
		[]byte(`{"report":null,"completeness_receipt":{}}`),
	} {
		if _, err := Compare(raw, bad); err == nil {
			t.Fatal("accepted malformed or ambiguous input")
		}
	}
	p, err := receiptprojection.Compile("delta.gooo", DeclarationSource(), "CompletenessDelta")
	if err != nil {
		t.Fatal(err)
	}
	for file, generate := range map[string]func() ([]byte, error){"delta.generated.go": p.Go, "delta.schema.json": p.JSONSchema} {
		got, err := generate()
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(file)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("generated projection differs", file, err)
		}
	}
	copy := DeclarationSource()
	copy[0] ^= 1
	if !declarationMatches() {
		t.Fatal("mutable declaration")
	}
}

func TestScopeChangesHaveStableEscapedPathsAndKeepNullPresence(t *testing.T) {
	a := map[string]any{"a/b": map[string]any{"~": nil}, "gone": nil}
	b := map[string]any{"a/b": map[string]any{"~": 1}, "new": nil}
	want := []string{"/scope/a~1b/~0", "/scope/gone", "/scope/new"}
	if got := changedScope(a, b); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
}
