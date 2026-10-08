package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestSourceSearchBudgetSurvivesGenerationAndReplay(t *testing.T) {
	ctx := context.Background()
	source, _ := sourceSearchReplayFixture(t)
	source = append(source, []byte(`
entity Note id "bodycodegen://note" fields {
    field text id "bodycodegen://note/text" type string required one
}
activity Describe(Note) -> Note computes "return input"
`)...)
	for _, budget := range []int{1, 3, 8, 16} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			changed := []byte(strings.Replace(string(source), `attempts "8"`, fmt.Sprintf(`attempts "%d"`, budget), 1))
			spec, err := SourceAssembly(ctx, "mixed.gooo", changed, "ClampNegativeToZero")
			if err != nil {
				t.Fatal(err)
			}
			r, err := GenerateWithSourceIRSearch(ctx, "mixed.gooo", changed, "ClampNegativeToZero", spec, "", "")
			if err != nil {
				t.Fatal(err)
			}
			s := r.Report.BodySearch
			if s.AttemptBudget == nil || *s.AttemptBudget != budget || s.AttemptedCandidates > budget || s.ProviderOperations != 0 {
				t.Fatal("source search lost its declared budget", s)
			}
			if err = VerifyIRBodySearchProjection(ctx, "mixed.gooo", changed, r); err != nil {
				t.Fatal("mixed record/search source did not replay", err)
			}
		})
	}
}

func TestSourceSearchBudgetAcceptsLegacyAndRejectsMismatches(t *testing.T) {
	ctx := context.Background()
	source, prior := sourceSearchReplayFixture(t)
	for _, budget := range []int{-1, 0, 1, 64} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			raw, _ := json.Marshal(prior)
			var changed Result
			if err := json.Unmarshal(raw, &changed); err != nil {
				t.Fatal(err)
			}
			changed.Report.BodySearch.AttemptBudget = &budget
			populateCompletenessReceipt(&changed.Report, "")
			if err := VerifyIRBodySearchProjection(ctx, "search.gooo", source, changed); err == nil || !strings.Contains(err.Error(), "attempt budget differs") {
				t.Fatal("search replay ignored the source budget", err)
			}
		})
	}
	prior.Report.BodySearch.AttemptBudget = nil
	populateCompletenessReceipt(&prior.Report, "")
	before, _ := json.Marshal(prior)
	if strings.Contains(string(before), `"attempt_budget"`) {
		t.Fatal("legacy omission changed the receipt shape")
	}
	if err := VerifyIRBodySearchProjection(ctx, "search.gooo", source, prior); err != nil {
		t.Fatal("legacy search replay failed", err)
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("legacy receipt was rewritten")
	}
}
