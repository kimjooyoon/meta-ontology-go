package packageruntime

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcepartition"
)

func TestSourceSplitChoiceMatchesMaterializedOutputs(t *testing.T) {
	manifest := splitFixture()
	observed, err := InspectSourceSplit(manifest, "split/ledger", "source.gooo")
	if err != nil {
		t.Fatal(err)
	}
	bestLines, bestID := observed.Input.Lines, ""
	for _, boundary := range observed.Candidates {
		partition, err := SplitSharedSource(manifest, "split/ledger", "source.gooo", boundary.BeforeID, "new.gooo")
		if err != nil {
			t.Fatal(err)
		}
		peak := 0
		for _, file := range partition.Files {
			lines := strings.Count(file.Source, "\n")
			if !strings.HasSuffix(file.Source, "\n") {
				lines++
			}
			peak = max(peak, lines)
		}
		if peak < bestLines || (peak == bestLines && boundary.BeforeID < bestID) {
			bestLines, bestID = peak, boundary.BeforeID
		}
	}
	choice, err := ChooseSourceSplit(manifest, "split/ledger", "source.gooo", 1)
	if err != nil || choice.State != "SELECTED" || choice.LimitSatisfied || choice.WithinLimit != 0 ||
		choice.Boundary == nil || choice.Boundary.BeforeID != bestID ||
		max(choice.Boundary.Left.Lines, choice.Boundary.Right.Lines) != bestLines {
		t.Fatal("choice differs from independently counted materialized outputs", choice, err)
	}
	if choice.InputSHA256 != observed.InputSHA256 || choice.InputImageDigest != observed.InputImageDigest {
		t.Fatal("choice lost the exact source and workspace identities")
	}
	if _, err := ChooseSourceSplit(manifest, "split/ledger", "source.gooo", 0); err == nil {
		t.Fatal("zero limit accepted")
	}
}

func TestSourceSplitChoiceCountsPartialProgressAndStableTies(t *testing.T) {
	observed := SourceSplitBoundaries{Input: sourcepartition.Size{Lines: 100}, Candidates: []SourceSplitBoundary{
		{BeforeID: "urn:z", Left: sourcepartition.Size{Lines: 55}, Right: sourcepartition.Size{Lines: 50}},
		{BeforeID: "urn:a", Left: sourcepartition.Size{Lines: 50}, Right: sourcepartition.Size{Lines: 55}},
		{BeforeID: "urn:no-reduction", Left: sourcepartition.Size{Lines: 100}, Right: sourcepartition.Size{Lines: 5}},
	}}
	choice := chooseSourceSplit(observed, 55)
	if choice.State != "SELECTED" || choice.Candidates != 3 || choice.Improving != 2 || choice.WithinLimit != 2 ||
		!choice.LimitSatisfied || choice.Boundary.BeforeID != "urn:a" {
		t.Fatal("incorrect candidate denominator or stable tie break", choice)
	}
	slices.Reverse(observed.Candidates)
	if replay := chooseSourceSplit(observed, 55); !reflect.DeepEqual(choice, replay) {
		t.Fatal("enumeration order changed the selected stable identity", replay)
	}
	partial := chooseSourceSplit(observed, 54)
	if partial.State != "SELECTED" || partial.LimitSatisfied || partial.WithinLimit != 0 {
		t.Fatal("intermediate progress was claimed to meet the requested limit", partial)
	}
	complete := chooseSourceSplit(observed, 100)
	if complete.State != "ALREADY_WITHIN_LIMIT" || !complete.LimitSatisfied || complete.Boundary != nil {
		t.Fatal("satisfied source received unnecessary split", complete)
	}
	observed.Candidates = observed.Candidates[:1]
	blocked := chooseSourceSplit(observed, 75)
	if blocked.State != "NO_IMPROVING_CUT" || blocked.LimitSatisfied || blocked.Boundary != nil || blocked.Improving != 0 {
		t.Fatal("header overhead was hidden by a non-improving split", blocked)
	}
}
