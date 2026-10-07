package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const dependencyRecordFixture = `package dependency
namespace dependency
entity Integer id "dependency://integer"
entity Box id "dependency://box" fields { field value id "dependency://value" type integer required one }
activity Seed(Integer) -> Box computes "return Box{value: input}"
activity Unrelated(Integer) -> Integer computes "return input"
activity Wrap(Integer) -> Box computes "let next = Seed(input); return Box{value: next.value * 2}" assembling {
    choice "value" field_value at "0" alternative "(next.value - 1) * 2" intent "Use the original input."
    value_case "[1]" -> "{\"value\":2}"
    value_case "[2]" -> "{\"value\":4}"
    attempts "2"
}
`

func TestRecordResumeRechecksAlternativeOnlyDependencyAtExhaustedBudget(t *testing.T) {
	source := []byte(`package alternate
namespace alternate
entity Integer id "alternate://integer"
entity Box id "alternate://box" fields { field value id "alternate://value" type integer required one }
activity Helper(Integer) -> Integer computes "return input + 1"
activity Wrap(Integer) -> Box computes "return Box{value: 0}" assembling {
    choice "value" field_value at "0" alternative "Helper(input)" intent "Use helper."
    value_case "[1]" -> "{\"value\":3}"
    value_case "[2]" -> "{\"value\":4}"
    attempts "2"
}
`)
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	prior, err := g.GenerateSourceAssembly(ctx, "alternative.gooo", source, "Wrap")
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Report.RecordAssembly.Attempts) != 2 || prior.Report.RecordAssembly.Passed != 0 {
		t.Fatal(prior.Report.RecordAssembly)
	}
	changed := []byte(strings.Replace(string(source), "return input + 1", "return input + 2", 1))
	result, err := ResumeRecordAssembly(ctx, "alternative.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Passed != 2 || r.SelectedMask != 1 || r.Continuation.RecheckedAttempts != 2 || r.Continuation.AddedAttempts != 0 {
		t.Fatal(r)
	}
	if _, err := RealizeSourceAssembly(ctx, "alternative.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordResumePreservesModelOriginAcrossDependencyChanges(t *testing.T) {
	text := strings.Replace(string(recordUpdatesFixture(t)), "let copy = input0", "let copy = Clone(input0)", 1)
	text += "\nactivity Clone(Candidate) -> Candidate computes `return input`\n"
	source := []byte(text)
	ctx := context.Background()
	g, err := NewTypedPathGenerator(writeSharedRecordContractModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	stop := fixedPolicy(t, "PROGRESS", "USE_OBSERVED_CANDIDATE")
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "origin.gooo", source, "Select", stop)
	if err != nil {
		t.Fatal(err)
	}
	if prior.Report.RecordAssembly.ModelCalls != 1 {
		t.Fatal("fixture did not predict")
	}
	// First retain another stage without changing source, then change the helper.
	prior, err = ResumeRecordAssembly(ctx, "origin.gooo", source, source, prior, stop)
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(text, "computes `return input`", "computes `let changed = input; changed.title = input.title + \"!\"; return changed`", 1))
	result, err := ResumeRecordAssembly(ctx, "origin.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r, old := result.Report.RecordAssembly, prior.Report.RecordAssembly
	if r.ModelCalls != 1 || r.Continuation.NewModelCalls != 0 || r.RankingSourceSHA256 != digest(source) ||
		r.Continuation.RecheckedAttempts != len(old.Attempts) || !reflect.DeepEqual(r.Context, old.Context) ||
		!reflect.DeepEqual(r.Prediction, old.Prediction) || !reflect.DeepEqual(r.Ranking, old.Ranking) {
		t.Fatal(r)
	}
	if _, err := RealizeSourceAssembly(ctx, "origin.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
	r.RankingSourceSHA256 = digest(changed)
	if _, err := RealizeSourceAssembly(ctx, "origin.gooo", changed, result); err == nil {
		t.Fatal("changed ranking origin replayed")
	}
}

func TestRecordResumeRechecksChangedDependencyBeforePolicy(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	source := []byte(dependencyRecordFixture)
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "dependency.gooo", source, "Wrap", assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(prior)
	changed := []byte(strings.Replace(string(source), "return Box{value: input}", "return Box{value: input + 1}", 1))
	result, err := ResumeRecordAssembly(ctx, "dependency.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Passed != 2 || r.SelectedMask != 1 || len(r.Attempts) != 2 || r.Attempts[0].Passed != 0 ||
		r.Continuation.RecheckedAttempts != 1 || r.Continuation.RetainedAttempts != 1 ||
		r.Continuation.AddedAttempts != 1 || r.Continuation.NewModelCalls != 0 ||
		r.RankingSourceSHA256 != digest(source) || r.ControlHistory[0].Source != string(source) {
		t.Fatal("old passing score survived a changed callee", r)
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("predecessor mutated")
	}
	if _, err := RealizeSourceAssembly(ctx, "dependency.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
	result.Report.RecordAssembly.Continuation.RecheckedAttempts = 0
	if _, err := RealizeSourceAssembly(ctx, "dependency.gooo", changed, result); err == nil {
		t.Fatal("changed recheck count replayed")
	}
}

func TestRecordResumeKeepsUnrelatedDependencyScores(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	source := []byte(dependencyRecordFixture)
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "dependency.gooo", source, "Wrap", assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(string(source), `computes "return input"`, `computes "return input + 9"`, 1))
	result, err := ResumeRecordAssembly(ctx, "dependency.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Continuation.RecheckedAttempts != 0 || r.Continuation.AddedAttempts != 0 || r.Passed != 2 {
		t.Fatal(r)
	}
	if _, err := RealizeSourceAssembly(ctx, "dependency.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
}
