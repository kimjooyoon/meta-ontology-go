package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const recordHoldoutSource = `package held
namespace held
entity Integer id "held://integer"
entity Box id "held://box" fields { field value id "held://box/value" type integer required one }
activity Choose(Integer) -> Box id "held://choose" computes "return Box{value: input}" assembling {
 choice "direction" field_value at "0" alternative "0 - input" intent "부호를 반대로 한다. Negate the input."
 value_case "[0]" -> "{\"value\":0}"
 holdout_value_case "[5]" -> "{\"value\":-5}"
 attempts "2"
}
`

func TestRecordHoldoutSourceCanBeConstructed(t *testing.T) {
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.GenerateSourceAssembly(context.Background(), "held.gooo", []byte(recordHoldoutSource), "Choose")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRecordHoldoutDoesNotSelectOrStop(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	a, err := g.GenerateSourceAssembly(ctx, "held.gooo", []byte(recordHoldoutSource), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(recordHoldoutSource, `holdout_value_case "[5]" -> "{\"value\":-5}"`, `holdout_value_case "[5]" -> "{\"value\":5}"`, 1)
	b, err := g.GenerateSourceAssembly(ctx, "held.gooo", []byte(changed), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	x, y := a.Report.RecordAssembly, b.Report.RecordAssembly
	if x.Holdout == nil || y.Holdout == nil || x.Holdout.Passed != 0 || y.Holdout.Passed != 1 ||
		x.Holdout.Total != 1 || x.Holdout.FieldsTotal != 1 || x.Holdout.AccuracyPercent == nil ||
		*x.Holdout.AccuracyPercent != 0 || x.Passed != 1 || x.Total != 1 || len(x.Attempts) != 1 ||
		x.SelectedMask != y.SelectedMask || !reflect.DeepEqual(x.Attempts, y.Attempts) ||
		!reflect.DeepEqual(x.Ranking, y.Ranking) || a.Source != b.Source || x.Holdout.SuiteSHA256 == y.Holdout.SuiteSHA256 {
		t.Fatal("held answers leaked into selection or lost their separate measurement", x, y)
	}
	for _, result := range []Result{a, b} {
		if result.Report.RecordAssembly.ModelCalls != 0 {
			t.Fatal("unexpected prediction")
		}
		d := recordAssemblyDimensions(result.Report.RecordAssembly)
		if len(d) != 4 || d[0].Numerator != 1 || d[0].Denominator != 1 || d[2].Denominator != 1 {
			t.Fatal(d)
		}
	}
}

func TestRecordHoldoutReplayBindsValuesAndJSON(t *testing.T) {
	ctx := context.Background()
	source := []byte(recordHoldoutSource)
	g, _ := NewTypedPathGenerator("")
	original, err := g.GenerateSourceAssembly(ctx, "held.gooo", source, "Choose")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	var decoded Result
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, err = RealizeSourceAssembly(ctx, "held.gooo", source, decoded); err != nil {
		t.Fatal("formatted held values failed replay", err)
	}
	mutations := []func(*RecordAssemblyReceipt){
		func(r *RecordAssemblyReceipt) { r.Holdout = nil },
		func(r *RecordAssemblyReceipt) { r.Holdout.Passed = 1 },
		func(r *RecordAssemblyReceipt) { r.Holdout.FieldsPassed = 1 },
		func(r *RecordAssemblyReceipt) { r.Holdout.Results[0].Actual = json.RawMessage(`{"value":-5}`) },
		func(r *RecordAssemblyReceipt) { r.Holdout.SelectedSourceSHA256 = "different" },
		func(r *RecordAssemblyReceipt) { r.Holdout.Cases[0].Expected = `{"value":5}` },
	}
	for i, mutate := range mutations {
		var value Result
		if err = json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		mutate(value.Report.RecordAssembly)
		if _, err = RealizeSourceAssembly(ctx, "held.gooo", source, value); err == nil {
			t.Fatalf("changed holdout %d replayed", i)
		}
	}
	if original.Report.RecordAssembly.Holdout.Passed != 0 {
		t.Fatal("verification changed the original")
	}
}

func TestRecordHoldoutEvaluationErrorKeepsSelectionAndUnknownScore(t *testing.T) {
	source := strings.Replace(recordHoldoutSource, "Box{value: input}", "Box{value: 12 / input}", 1)
	source = strings.Replace(source, `value_case "[0]" -> "{\"value\":0}"`, `value_case "[3]" -> "{\"value\":4}"`, 1)
	source = strings.Replace(source, `holdout_value_case "[5]"`, `holdout_value_case "[0]"`, 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "held.gooo", []byte(source), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Passed != 1 || r.SelectedMask != 0 || r.Holdout.Error == "" || r.Holdout.AccuracyPercent != nil || r.Holdout.Total != 1 {
		t.Fatal(r)
	}
	for _, d := range recordHoldoutDimensions(r.Holdout) {
		if d.Status != "UNKNOWN" {
			t.Fatal(d)
		}
	}
	if _, err = RealizeSourceAssembly(context.Background(), "held.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordHoldoutContinuationEvaluatesChangedHelper(t *testing.T) {
	source := strings.Replace(recordHoldoutSource, "Box{value: input}", "Box{value: Helper(input)}", 1) + "\nactivity Helper(Integer) -> Integer computes \"return input\"\n"
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	prior, err := g.GenerateSourceAssembly(ctx, "held.gooo", []byte(source), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	old, _ := json.Marshal(prior)
	changed := strings.Replace(source, `computes "return input"`, `computes "return 0 - input"`, 1)
	result, err := ResumeRecordAssembly(ctx, "held.gooo", []byte(source), []byte(changed), prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Continuation.RetainedAttempts != 1 || r.Continuation.RecheckedAttempts != 1 || r.Continuation.AddedAttempts != 0 ||
		r.Continuation.NewModelCalls != 0 || r.SelectedMask != 0 || prior.Report.RecordAssembly.Holdout.Passed != 0 || r.Holdout.Passed != 1 {
		t.Fatal(r)
	}
	if _, err = RealizeSourceAssembly(ctx, "held.gooo", []byte(changed), result); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(prior)
	if !bytes.Equal(old, after) {
		t.Fatal("continuation rewrote prior evaluation")
	}
}

func TestRecordHoldoutRunsOnlyForCompletedLocalSelection(t *testing.T) {
	ctx := context.Background()
	source := []byte(recordHoldoutSource)
	plan, err := prepareRecordAssembly(ctx, "held.gooo", source, "Choose")
	if err != nil {
		t.Fatal(err)
	}
	receipt := newRecordAssemblyReceipt(source, plan)
	if err = searchRecordAssembly(ctx, plan, receipt); err != nil {
		t.Fatal(err)
	}
	candidate, err := emitRecordAssembly(ctx, "held.gooo", source, plan, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Holdout != nil || candidate.Report.RecordAssembly.Holdout != nil {
		t.Fatal("candidate projection acquired held outcomes")
	}
	selected, err := emitSelectedRecordAssembly(ctx, "held.gooo", source, plan, receipt)
	if err != nil || selected.Report.RecordAssembly.Holdout == nil {
		t.Fatal("completed selection lost held evaluation", err)
	}
}

func TestRecordHoldoutPreservesOptionalPresenceAndExactIntegerValues(t *testing.T) {
	source := []byte(`package values
namespace values
entity Request id "values://request" fields {
 field label id "values://request/label" type string required one
 field amount id "values://request/amount" type integer required one
}
entity Result id "values://result" fields {
 field label id "values://result/label" type string required one
 field total id "values://result/total" type integer optional one
}
activity Build(Request) -> Result computes "let amount = input.amount; return Result{label: input.label, total: nil}" assembling {
 choice "total" field_value at "1" alternative "&amount" intent "Preserve the supplied amount, including zero."
 value_case "[{\"label\":\"seed\",\"amount\":0}]" -> "{\"label\":\"seed\",\"total\":0}"
 holdout_value_case "[{\"label\":\"한글\",\"amount\":9007199254740993}]" -> "{\"label\":\"한글\",\"total\":9007199254740993}"
 holdout_value_case "[{\"label\":\"English\",\"amount\":0}]" -> "{\"label\":\"English\"}"
 attempts "2"
}
`)
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(ctx, "values.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.SelectedMask != 1 || r.Passed != 1 || r.Holdout.Passed != 1 || r.Holdout.Total != 2 || r.Holdout.FieldsPassed != 3 || r.Holdout.FieldsTotal != 4 {
		t.Fatal(r)
	}
	if !strings.Contains(string(r.Holdout.Results[0].Actual), "9007199254740993") {
		t.Fatal("integer precision was lost")
	}
	field := r.Holdout.Results[1].Fields[1]
	if field.ExpectedPresent == nil || *field.ExpectedPresent || field.ActualPresent == nil || !*field.ActualPresent || field.Passed {
		t.Fatal("absent expected value became a present zero", field)
	}
	if _, err = RealizeSourceAssembly(ctx, "values.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordHoldoutRejectsSameIntegerWithDifferentSpelling(t *testing.T) {
	source := []byte(strings.Replace(recordHoldoutSource, `holdout_value_case "[5]"`, `holdout_value_case "[-0]"`, 1))
	g, _ := NewTypedPathGenerator("")
	_, err := g.GenerateSourceAssembly(context.Background(), "duplicate.gooo", source, "Choose")
	if err == nil || !strings.Contains(err.Error(), "overlap") {
		t.Fatal("selection 0 and held -0 are the same input", err)
	}
}
