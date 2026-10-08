package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// Each individual alternative retains one use of saved. Mask 3 removes both
// reads, while mask 6 completes the finite contract. An unused saved local still
// evaluates its initializer and is now a valid Gooo binding.
func recordUnusedCombinationFixture(t *testing.T) []byte {
	t.Helper()
	source := string(recordUpdatesFixture(t))
	start, end := strings.Index(source, "computes `"), strings.Index(source, "` assembling {")
	if start < 0 || end <= start {
		t.Fatal("fixture shape differs")
	}
	body := "computes `let copy = input0\nlet saved = copy\nif input1 && copy.state != \"ready\" {\n" +
		"    copy.title = saved.title\n    copy.state = saved.state\n    copy.reason = \"deferred\"\n}\nreturn copy"
	return []byte(source[:start] + body + source[end:])
}

func TestRecordAssemblyTypedRejectionBeforeAnyValidCandidate(t *testing.T) {
	ctx := context.Background()
	source := recordUnusedCombinationFixture(t)
	plan, err := prepareRecordAssembly(ctx, "combination.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	// Inject a real field-type mismatch after preflight to exercise rejection
	// independently of the source language's permitted unused local bindings.
	plan.sites[0].choice.Second = "true"
	r := newRecordAssemblyReceipt(source, plan)
	r.Ranking = []uint16{3, 6}
	if err = searchRecordAssembly(ctx, plan, r); err != nil || len(r.Attempts) != 2 || r.SelectedMask != 6 ||
		r.Attempts[0].Status != "TYPECHECK_FAILED" || r.FieldsPassed != 15 {
		t.Fatal("invalid first proposal prevented continuation", err)
	}
	plan.spec.MaxAttempts = 1
	r = newRecordAssemblyReceipt(source, plan)
	r.Ranking = []uint16{3, 6}
	if err = searchRecordAssembly(ctx, plan, r); err == nil || !strings.Contains(err.Error(), "no valid typed candidate") ||
		len(r.Attempts) != 1 || len(r.Cases) != 0 {
		t.Fatal("invalid-only budget acquired a valid result", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err = searchRecordAssembly(cancelled, plan, newRecordAssemblyReceipt(source, plan)); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled request continued", err)
	}
	plan.sites[0].choice.Second = "unsupported()"
	r = newRecordAssemblyReceipt(source, plan)
	r.Ranking = []uint16{1}
	if err = searchRecordAssembly(ctx, plan, r); err == nil || len(r.Attempts) != 0 {
		t.Fatal("non-type generation error was swallowed", err)
	}
}

func TestRecordAssemblyModelPredictsOnceAcrossUnusedLocalCandidate(t *testing.T) {
	g, err := NewTypedPathGenerator(writeOriginRecordModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	source := recordUnusedCombinationFixture(t)
	result, err := g.GenerateSourceAssembly(context.Background(), "model.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	scored := 0
	for _, attempt := range r.Attempts {
		if attempt.Mask == 3 && attempt.Status == "" && attempt.Total == 5 {
			scored++
		}
	}
	if r.ModelCalls != 1 || r.FieldsPassed != 15 || r.Context.Status != "ENCODED" || scored != 1 {
		t.Fatalf("single prediction across unused local: calls=%d fields=%d scored=%d", r.ModelCalls, r.FieldsPassed, scored)
	}
	replay, err := RealizeSourceAssembly(context.Background(), "model.gooo", source, result)
	if err != nil || replay.ModelCalls != 0 {
		t.Fatal("unused local required a new prediction on replay", err)
	}
}

func TestRecordConstructorCombinationAlsoContinues(t *testing.T) {
	source := string(recordUnusedCombinationFixture(t))
	start, end := strings.Index(source, "computes `"), strings.Index(source, "` assembling {")
	body := "computes `let copy = input0\nlet saved = copy\nif input1 && copy.state != \"ready\" {\n" +
		"    return Candidate{title: saved.title, state: saved.state, reason: \"deferred\"}\n}\nreturn copy"
	source = source[:start] + body + source[end:]
	source = strings.ReplaceAll(source, "field_update", "field_value")
	source = strings.Replace(source, `alternative "copy.title + \":\" + copy.state"`, `alternative "copy.title + \":ready\""`, 1)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "constructor.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal("constructor combination stopped finite selection", err)
	}
	r := result.Report.RecordAssembly
	if r.SelectedMask != 6 || r.FieldsPassed != 15 || len(r.Attempts) != 7 || r.Attempts[3].Status != "" ||
		r.Attempts[3].Total != 5 {
		t.Fatalf("constructor continuation: mask=%d fields=%d attempts=%d", r.SelectedMask, r.FieldsPassed, len(r.Attempts))
	}
	if _, err = RealizeSourceAssembly(context.Background(), "constructor.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordSuccessfulAttemptKeepsPreviousJSONShape(t *testing.T) {
	raw, err := json.Marshal(RecordAssemblyAttempt{Mask: 1, Passed: 2, Total: 5, FieldsPassed: 9, FieldsTotal: 15})
	if err != nil || string(raw) != `{"mask":1,"passed":2,"total":5,"fields_passed":9,"fields_total":15}` {
		t.Fatal("successful attempt JSON shape changed", string(raw), err)
	}
}

func TestRecordAssemblyScoresUnusedLocalCombination(t *testing.T) {
	source := recordUnusedCombinationFixture(t)
	ctx := context.Background()
	if err := ValidateSourceAssembly(ctx, "combination.gooo", source, "Select"); err != nil {
		t.Fatal("each alternative should pass preflight", err)
	}
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "combination.gooo", source, "Select")
	if err != nil {
		t.Fatal("unused local stopped a later complete candidate", err)
	}
	r := result.Report.RecordAssembly
	if r.SelectedMask != 6 || r.Status != "COMPLETE_FINITE" || r.FieldsPassed != 15 || r.FieldsTotal != 15 ||
		len(r.Attempts) != 7 || r.ModelCalls != 0 || !result.Report.TypecheckPassed {
		t.Fatalf("finite continuation differs: %+v", r)
	}
	partial := r.Attempts[3]
	if partial.Mask != 3 || partial.Status != "" || partial.Reason != "" ||
		partial.Passed != 2 || partial.Total != 5 || partial.FieldsPassed != 12 || partial.FieldsTotal != 15 {
		t.Fatalf("unused-local combination was not scored: %+v", partial)
	}
	realized, err := RealizeSourceAssembly(ctx, "combination.gooo", source, result)
	if err != nil || realized.ModelCalls != 0 || realized.Source != result.GoooSource {
		t.Fatal("partial attempt did not replay", err)
	}
	next, err := g.GenerateSourceAssembly(ctx, "combination.gooo", []byte(result.GoooSource), "Select")
	if err != nil || next.Source != result.Source || next.GoooSource != result.GoooSource {
		t.Fatal("checkpoint fixed point differs", err)
	}
	for _, mutate := range []func(*Result){
		func(v *Result) { v.Report.RecordAssembly.Attempts[3].Reason = "other error" },
		func(v *Result) { v.Report.RecordAssembly.Attempts[3].Status = "TYPECHECK_FAILED" },
		func(v *Result) { v.Report.RecordAssembly.Attempts[3].Total = 0 },
	} {
		raw, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		var changed Result
		if err = json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		mutate(&changed)
		if _, err = RealizeSourceAssembly(ctx, "combination.gooo", source, changed); err == nil {
			t.Fatal("changed scored-attempt observation was accepted")
		}
	}
}

func TestRecordAssemblyUnusedCombinationConsumesBudgetAndKeepsPartial(t *testing.T) {
	source := []byte(strings.Replace(string(recordUnusedCombinationFixture(t)), "attempts \"8\"", "attempts \"4\"", 1))
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "partial.gooo", source, "Select")
	if err != nil {
		t.Fatal("partial valid candidate was lost", err)
	}
	r := result.Report.RecordAssembly
	if len(r.Attempts) != 4 || r.Attempts[3].Status != "" || r.Attempts[3].Total != 5 || r.SelectedMask != 2 ||
		r.Status != "PARTIAL_FINITE" || r.FieldsPassed != 12 || r.FieldsTotal != 15 {
		t.Fatalf("bounded partial continuation differs: %+v", r)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "partial.gooo", source, result); err != nil {
		t.Fatal("partial unused-local attempt did not replay", err)
	}
}
