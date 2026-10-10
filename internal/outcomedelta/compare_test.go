package outcomedelta

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

func delivery(input, actual, expected string) bodyexecution.CompositionDelivery {
	d := bodyexecution.CompositionDelivery{ActivityID: "queue://activity/main", Input: json.RawMessage(input), Actual: json.RawMessage(actual), Expected: json.RawMessage(expected)}
	if input == "" {
		d.Input = nil
	}
	if actual == "" {
		d.Actual = nil
	}
	if expected == "" {
		d.Expected = nil
	}
	if expected != "" && actual != "" {
		_, a, _ := canonical([]byte(actual))
		_, e, _ := canonical([]byte(expected))
		pass := a == e
		d.Passed = &pass
	}
	return d
}

func runtimeFor(rows ...[]bodyexecution.CompositionDelivery) bodyexecution.CompositionRuntime {
	r := bodyexecution.CompositionRuntime{Schema: "gooo/body-composition-runtime/v1", Stage: "COMPLETE", ProducerSourceSHA: "recorded-source"}
	for i, row := range rows {
		r.Traces = append(r.Traces, bodyexecution.CompositionTrace{CaseIndex: i, Deliveries: row})
		for _, d := range row {
			if len(d.Expected) > 0 {
				r.FiniteTotal++
			}
			if d.Passed != nil {
				if *d.Passed {
					r.FinitePassed++
				}
			}
		}
	}
	return r
}

func raw(t *testing.T, value any) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func compare(t *testing.T, b, a bodyexecution.CompositionRuntime) *OutcomeDelta {
	t.Helper()
	r, err := Compare(raw(t, b), raw(t, a))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestOutcomeAssessment(t *testing.T) {
	for _, test := range []struct{ name, oldActual, oldExpected, newActual, newExpected, outcome, requirement, assessment string }{
		{"same", "1", "1", "1", "1", "UNCHANGED", "UNCHANGED", "STILL_MATCHING"},
		{"goal", "1", "1", "2", "2", "CHANGED", "CHANGED", "REQUIREMENT_CHANGED"},
		{"goal-only", "1", "1", "1", "2", "UNCHANGED", "CHANGED", "REQUIREMENT_CHANGED"},
		{"regression", "1", "1", "2", "1", "CHANGED", "UNCHANGED", "REGRESSION"},
		{"improvement", "2", "1", "1", "1", "CHANGED", "UNCHANGED", "IMPROVEMENT"},
		{"still-failing", "2", "1", "3", "1", "CHANGED", "UNCHANGED", "STILL_FAILING"},
		{"no-expectations", "1", "", "2", "", "CHANGED", "UNOBSERVED", "UNOBSERVED"},
		{"new-expectation", "1", "", "1", "1", "UNCHANGED", "UNOBSERVED", "UNOBSERVED"},
		{"explicit-null", "null", "null", "null", "null", "UNCHANGED", "UNCHANGED", "STILL_MATCHING"},
	} {
		t.Run(test.name, func(t *testing.T) {
			b := runtimeFor([]bodyexecution.CompositionDelivery{delivery("1", test.oldActual, test.oldExpected)})
			a := runtimeFor([]bodyexecution.CompositionDelivery{delivery("1", test.newActual, test.newExpected)})
			r := compare(t, b, a)
			row := r.Rows[0]
			if row.OutcomeChange != test.outcome || row.RequirementChange != test.requirement || row.Assessment != test.assessment {
				t.Fatalf("%+v", row)
			}
			for _, n := range r.ComparatorOperations {
				if n != 0 {
					t.Fatal("comparison executed work")
				}
			}
		})
	}
}

func TestNoOutputObserved(t *testing.T) {
	before := runtimeFor([]bodyexecution.CompositionDelivery{delivery("1", "1", "1")})
	after := []byte(`{"schema":"gooo/body-composition-runtime/v1","stage":"EXECUTE_1","finite_total":1,"traces":[{"case_index":0,"deliveries":[{"activity_id":"queue://activity/main","input":1,"expected":1}]}]}`)
	r, err := Compare(raw(t, before), after)
	if err != nil || r.Rows[0].OutcomeChange != "UNOBSERVED" || r.Rows[0].Assessment != "UNOBSERVED" {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestExactIntegersAndCaseOrder(t *testing.T) {
	b := runtimeFor([]bodyexecution.CompositionDelivery{delivery("9007199254740992", "9007199254740992", "9007199254740992")},
		[]bodyexecution.CompositionDelivery{delivery("9007199254740993", "9007199254740993", "9007199254740993")})
	a := runtimeFor([]bodyexecution.CompositionDelivery{delivery("9007199254740993", "9007199254740992", "9007199254740993")},
		[]bodyexecution.CompositionDelivery{delivery("9007199254740992", "9007199254740992", "9007199254740992")})
	r := compare(t, b, a)
	if len(r.Rows) != 2 || r.Counts["assessment:REGRESSION"] != 1 || r.Counts["assessment:STILL_MATCHING"] != 1 {
		t.Fatalf("%+v", r)
	}
	if !bytes.Contains(raw(t, r), []byte("9007199254740993")) {
		t.Fatal("large integer lost")
	}
	if !bytes.Equal(raw(t, r), raw(t, compare(t, b, a))) {
		t.Fatal("comparison is not deterministic")
	}
}

func TestWholeCallerTupleAndBoundInputs(t *testing.T) {
	root := delivery("1", "2", "2")
	root.ActivityID = "flow://producer"
	child := delivery("2", "3", "3")
	child.ActivityID = "flow://consumer"
	child.ProducerID = root.ActivityID
	nextRoot := delivery("1", "3", "3")
	nextRoot.ActivityID = root.ActivityID
	nextChild := delivery("3", "4", "4")
	nextChild.ActivityID = child.ActivityID
	nextChild.ProducerID = root.ActivityID
	r := compare(t, runtimeFor([]bodyexecution.CompositionDelivery{root, child}), runtimeFor([]bodyexecution.CompositionDelivery{nextChild, nextRoot}))
	if len(r.Rows) != 2 || r.Counts["presence:BOTH"] != 2 || r.Counts["requirement:CHANGED"] != 2 {
		t.Fatalf("%+v", r)
	}
	// Independent root B changes. Equal root A alone cannot align the complete case.
	child.ProducerID = ""
	nextChild.ProducerID = ""
	r = compare(t, runtimeFor([]bodyexecution.CompositionDelivery{root, child}), runtimeFor([]bodyexecution.CompositionDelivery{root, nextChild}))
	if r.Counts["presence:BOTH"] != 0 || r.Counts["presence:BEFORE_ONLY"] != 2 || r.Counts["presence:AFTER_ONLY"] != 2 {
		t.Fatalf("%+v", r)
	}
}

func TestPortsRecordsAndDuplicates(t *testing.T) {
	d := delivery("", "true", "true")
	d.Inputs = []bodyexecution.CompositionPortDelivery{{Port: "input1", Value: json.RawMessage(`{"x":9007199254740993,"y":2}`)}, {Port: "input0", Value: json.RawMessage(`false`)}}
	e := d
	e.Inputs = []bodyexecution.CompositionPortDelivery{{Port: "input0", Value: json.RawMessage(`false`)}, {Port: "input1", Value: json.RawMessage(`{"y":2,"x":9007199254740993}`)}}
	r := compare(t, runtimeFor([]bodyexecution.CompositionDelivery{d}, []bodyexecution.CompositionDelivery{d}), runtimeFor([]bodyexecution.CompositionDelivery{e}))
	if len(r.Rows) != 1 || len(r.Rows[0].Before) != 2 || r.Rows[0].Assessment != "STILL_MATCHING" {
		t.Fatalf("%+v", r)
	}
	e.Actual = json.RawMessage(`false`)
	no := false
	e.Passed = &no
	r = compare(t, runtimeFor([]bodyexecution.CompositionDelivery{d}, []bodyexecution.CompositionDelivery{e}), runtimeFor([]bodyexecution.CompositionDelivery{d}))
	if len(r.Rows) != 1 || r.Rows[0].Assessment != "AMBIGUOUS" {
		t.Fatalf("%+v", r)
	}
}

func TestFaultBlockedAndPartialObservation(t *testing.T) {
	ok := delivery("1", "1", "1")
	fault := delivery("1", "", "1")
	no := false
	fault.Passed = &no
	fault.Fault = &bodyexecution.CompositionFault{Kind: "division_by_zero", Site: bodyexecution.CompositionFaultSite{Operator: "/", ExpressionID: "expression-1"}, Left: 1}
	r := compare(t, runtimeFor([]bodyexecution.CompositionDelivery{ok}), runtimeFor([]bodyexecution.CompositionDelivery{fault}))
	if r.Rows[0].Assessment != "REGRESSION" || r.Rows[0].After[0]["state"] != "FAULT" {
		t.Fatalf("%+v", r)
	}
	blocked := fault
	blocked.Fault = nil
	blocked.BlockedBy = []string{"producer"}
	r = compare(t, runtimeFor([]bodyexecution.CompositionDelivery{blocked}), runtimeFor([]bodyexecution.CompositionDelivery{ok}))
	if r.Rows[0].Assessment != "IMPROVEMENT" {
		t.Fatalf("%+v", r)
	}
	partial := runtimeFor()
	partial.Stage = "BUILD"
	partial.FiniteTotal = 10
	r = compare(t, runtimeFor([]bodyexecution.CompositionDelivery{ok}), partial)
	if r.Rows[0].Presence != "BEFORE_ONLY" || r.Rows[0].Assessment != "UNOBSERVED" || r.AfterRuntime["stage"] != "BUILD" {
		t.Fatalf("%+v", r)
	}
}

func TestSupportedEnvelopesAndLatestOnly(t *testing.T) {
	runtime := runtimeFor([]bodyexecution.CompositionDelivery{delivery("1", "2", "2")})
	evaluation := bodyexecution.JointEvaluation{Runtime: runtime}
	for _, envelope := range []any{runtime, evaluation,
		map[string]any{"generated_now": true, "construction": bodyexecution.JointConstruction{}, "evaluation": evaluation},
		map[string]any{"composition": bodyexecution.Composition{}, "runtime": runtime, "runtime_history": []bodyexecution.CompositionRuntime{runtime, runtime}},
	} {
		r, err := Compare(raw(t, envelope), raw(t, runtime))
		if err != nil || len(r.Rows) != 1 || r.Counts["assessment:STILL_MATCHING"] != 1 {
			t.Fatalf("%+v %v", r, err)
		}
	}
}

func TestInvalidAndConflictingRecords(t *testing.T) {
	base := runtimeFor([]bodyexecution.CompositionDelivery{delivery("1", "1", "1")})
	for _, mutate := range []func(*bodyexecution.CompositionRuntime){
		func(r *bodyexecution.CompositionRuntime) { r.Schema = "unknown" },
		func(r *bodyexecution.CompositionRuntime) { r.FinitePassed++ },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Deliveries[0].ActivityID = "" },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Deliveries[0].Input = nil },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Deliveries[0].Input = json.RawMessage(`1.1`) },
		func(r *bodyexecution.CompositionRuntime) {
			r.Traces[0].Deliveries[0].Input = json.RawMessage(`9223372036854775808`)
		},
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Deliveries[0].Expected = json.RawMessage(`2`) },
		func(r *bodyexecution.CompositionRuntime) { r.Traces[0].Deliveries[0].ProducerID = "absent" },
		func(r *bodyexecution.CompositionRuntime) { r.Traces = append(r.Traces, r.Traces[0]) },
	} {
		var changed bodyexecution.CompositionRuntime
		if err := json.Unmarshal(raw(t, base), &changed); err != nil {
			t.Fatal(err)
		}
		mutate(&changed)
		if _, err := Compare(raw(t, base), raw(t, changed)); err == nil {
			t.Fatal("accepted inconsistent record")
		}
	}
	for _, invalid := range []string{`{"schema":"x","schema":"y"}`, `null`, string(raw(t, base)) + ` {}`, strings.Replace(string(raw(t, base)), `"stage":`, `"Stage":`, 1)} {
		if _, err := Compare(raw(t, base), []byte(invalid)); err == nil {
			t.Fatal("accepted invalid JSON")
		}
	}
}

func TestSourceOwnedProjection(t *testing.T) {
	p, err := receiptprojection.Compile("delta.gooo", declaration, "OutcomeDelta")
	if err != nil {
		t.Fatal(err)
	}
	goBytes, err := p.Go()
	if err != nil {
		t.Fatal(err)
	}
	schemaBytes, err := p.JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string][]byte{"delta.generated.go": goBytes, "delta.schema.json": schemaBytes} {
		got, err := os.ReadFile(path)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("projection differs: %s %v", path, err)
		}
	}
}
