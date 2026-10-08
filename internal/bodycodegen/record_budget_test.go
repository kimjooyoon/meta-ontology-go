package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRecordBudgetReportsSourceLimitSeparatelyFromRanking(t *testing.T) {
	ctx := context.Background()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	for _, budget := range []int{1, 3, 8, 16} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			source := []byte(strings.Replace(string(recordUpdatesFixture(t)), `attempts "8"`, fmt.Sprintf(`attempts "%d"`, budget), 1))
			result, err := g.GenerateSourceAssembly(ctx, "record.gooo", source, "Select")
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(result.Report.RecordAssembly)
			var receipt map[string]any
			if err = json.Unmarshal(raw, &receipt); err != nil || receipt["attempt_budget"] != float64(budget) || len(result.Report.RecordAssembly.Ranking) != 8 {
				t.Fatal("source budget missing or replaced with candidate count", receipt["attempt_budget"], err)
			}
			if _, err = RealizeSourceAssembly(ctx, "record.gooo", source, result); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRecordBudgetReplayChecksSourceAndAcceptsLegacyOmission(t *testing.T) {
	ctx := context.Background()
	source := recordUpdatesFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "record.gooo", source, "Select", fixedPolicy(t, "PROGRESS", "USE_OBSERVED_CANDIDATE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []int{0, -1, 1, 64} {
		t.Run(fmt.Sprint(bad), func(t *testing.T) {
			raw, _ := json.Marshal(prior)
			var changed Result
			if err := json.Unmarshal(raw, &changed); err != nil {
				t.Fatal(err)
			}
			changed.Report.RecordAssembly.AttemptBudget = &bad
			populateCompletenessReceipt(&changed.Report, "")
			if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, changed); err == nil || !strings.Contains(err.Error(), "attempt budget differs") {
				t.Fatal("a recorded limit different from source replayed", err)
			}
		})
	}
	prior.Report.RecordAssembly.AttemptBudget = nil
	populateCompletenessReceipt(&prior.Report, "")
	before, _ := json.Marshal(prior)
	if strings.Contains(string(before), `"attempt_budget"`) {
		t.Fatal("legacy omission serialized a new field")
	}
	if _, err = RealizeSourceAssembly(ctx, "record.gooo", source, prior); err != nil {
		t.Fatal("legacy replay", err)
	}
	continued, err := ResumeRecordAssembly(ctx, "record.gooo", source, source, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal("legacy continuation", err)
	}
	r := continued.Report.RecordAssembly
	if r.AttemptBudget == nil || *r.AttemptBudget != 8 || r.Continuation.NewModelCalls != 0 {
		t.Fatal("continued legacy record did not recover the source budget", r.AttemptBudget)
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("legacy predecessor was rewritten")
	}
	if _, err = RealizeSourceAssembly(ctx, "record.gooo", source, continued); err != nil {
		t.Fatal(err)
	}
}
