package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRecordResumeKeepsPrefixAndRemainingBudget(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	source := recordUpdatesFixture(t)
	stop := fixedPolicy(t, "PROGRESS", "USE_OBSERVED_CANDIDATE")
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "record.gooo", source, "Select", stop)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(prior)
	still, err := ResumeRecordAssembly(ctx, "record.gooo", source, source, prior, stop)
	if err != nil || still.Report.RecordAssembly.Continuation.AddedAttempts != 0 {
		t.Fatal("new Gooo policy did not retain a pause", err)
	}
	result, err := ResumeRecordAssembly(ctx, "record.gooo", source, source, still, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Passed != 5 || r.FieldsPassed != 15 || len(r.Attempts) != 8 || len(r.ControlHistory) != 2 ||
		r.Continuation.RetainedAttempts != 1 || r.Continuation.AddedAttempts != 7 || r.Continuation.NewModelCalls != 0 ||
		r.Control.Entry == nil || !r.Control.Entry.Continue || !reflect.DeepEqual(r.Attempts[:1], prior.Report.RecordAssembly.Attempts) {
		t.Fatal("continuation replaced the prefix or reset the budget", r)
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("resume changed its predecessor")
	}
	raw, _ := json.Marshal(result)
	var decoded Result
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, decoded); err != nil {
		t.Fatal(err)
	}
	decoded.Report.RecordAssembly.ControlHistory[0].Attempts++
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, decoded); err == nil {
		t.Fatal("changed prefix replayed")
	}
}

func TestRecordResumeUsesCapturedModelWithoutAnotherCall(t *testing.T) {
	ctx := context.Background()
	g, err := NewTypedPathGenerator(writeOriginRecordModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	source := recordUpdatesFixture(t)
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "record.gooo", source, "Select", fixedPolicy(t, "PROGRESS", "USE_OBSERVED_CANDIDATE"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ResumeRecordAssembly(ctx, "record.gooo", source, source, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 1 || r.Continuation.NewModelCalls != 0 || !reflect.DeepEqual(r.Ranking, prior.Report.RecordAssembly.Ranking) ||
		!reflect.DeepEqual(r.Prediction, prior.Report.RecordAssembly.Prediction) || r.FieldsPassed != 15 {
		t.Fatal("saved model order changed", r)
	}
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	r.Model.MetadataSHA256 = "changed"
	if prior.Report.RecordAssembly.Model.MetadataSHA256 == "changed" {
		t.Fatal("model observation aliases predecessor")
	}
}

func TestRecordResumeCannotCreateBudgetOrChangeTarget(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	source := []byte(strings.Replace(string(recordUpdatesFixture(t)), `attempts "8"`, `attempts "1"`, 1))
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "record.gooo", source, "Select", assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ResumeRecordAssembly(ctx, "record.gooo", source, source, prior, fixedPolicy(t, "PROGRESS", "CONTINUE_CANDIDATES"))
	if err != nil || len(result.Report.RecordAssembly.Attempts) != 1 || result.Report.RecordAssembly.Continuation.AddedAttempts != 0 {
		t.Fatal("exhausted budget restarted", err)
	}
	changed := []byte(strings.Replace(string(source), `attempts "1"`, `attempts "8"`, 1))
	if _, err := ResumeRecordAssembly(ctx, "record.gooo", source, changed, prior, assemblyPolicyFixture(t)); err == nil {
		t.Fatal("changed obligations continued under old evidence")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := ResumeRecordAssembly(cancelled, "record.gooo", source, source, prior, assemblyPolicyFixture(t)); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled continuation lost cancellation", err)
	}
}

func TestRecordResumeBoundsHistoryAndReplaysEntry(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	source := recordUpdatesFixture(t)
	stop := fixedPolicy(t, "PROGRESS", "USE_OBSERVED_CANDIDATE")
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "record.gooo", source, "Select", stop)
	if err != nil {
		t.Fatal(err)
	}
	for range recordControlStageLimit {
		prior, err = ResumeRecordAssembly(ctx, "record.gooo", source, source, prior, stop)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ResumeRecordAssembly(ctx, "record.gooo", source, source, prior, stop); err == nil {
		t.Fatal("history exceeded its declared bound")
	}
	prior.Report.RecordAssembly.Control.Entry.Continue = true
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, prior); err == nil {
		t.Fatal("changed entry decision replayed")
	}
}
