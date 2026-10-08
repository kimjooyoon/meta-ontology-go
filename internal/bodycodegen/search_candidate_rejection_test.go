package bodycodegen

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSourceSearchCandidateRetainsUnscoredRejection(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-search-rejection/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, localEvaluation := range []bool{false, true} {
		current := string(source)
		if localEvaluation {
			current = strings.Replace(current, "if input == 0 { return 0 }; return input / __GOOO_BODY_HOLE_value__", "let denominator = __GOOO_BODY_HOLE_value__; return input / denominator", 1)
			current = strings.Replace(current, `case "0" -> "0"`, `case "1" -> "0"`, 1)
		}
		set, err := PlanSourceSearchCandidates(context.Background(), "rejection.gooo", []byte(current), "Choose")
		if err != nil {
			t.Fatal(err)
		}
		candidate := ""
		for _, c := range set.Candidates {
			if c.Expression == "0" {
				candidate = c.ID
			}
		}
		r, selected, err := RealizeSourceSearchCandidate(context.Background(), "rejection.gooo", []byte(current), "Choose", candidate)
		var rejected *SourceSearchCandidateRejection
		if !errors.As(err, &rejected) || r.Attempt.Error != err.Error() || r.Attempt.TypecheckPassed != localEvaluation ||
			r.Attempt.ScoringCompleted || r.Attempt.AccuracyPercent != nil || r.Attempt.TestCasesPassed != 0 || r.Attempt.TestCasesTotal != 1 ||
			len(r.Attempt.CaseResults) != 0 || len(selected) != 0 || r.SelectedSourceSHA256 != "" || r.InputSourceSHA256 == "" || r.PlanSHA256 == "" {
			t.Fatal(r, err)
		}
	}
}

func TestSourceSearchRequestFailuresAreNotCandidateRejections(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-search-rejection/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, ctx := range []context.Context{nil, canceled, context.Background()} {
		_, _, err := RealizeSourceSearchCandidate(ctx, "rejection.gooo", source, "Choose", "unknown")
		var rejected *SourceSearchCandidateRejection
		if err == nil || errors.As(err, &rejected) {
			t.Fatal("request failure became a rejected candidate", err)
		}
	}
}
