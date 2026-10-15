package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const jointSearchCheckpointSource = `package searchcheckpoint
namespace searchcheckpoint
entity Integer id "searchcheckpoint://integer"
entity Box id "searchcheckpoint://box" fields {
 field value id "searchcheckpoint://value" type integer required one
}

activity Seed(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "input * 2" intent "Double the seed."
 value_case "[3]" -> "{\"value\":6}"
 attempts "2"
}
activity Build(Integer) -> Integer computes "let box = Seed(input + __GOOO_BODY_HOLE_offset__); return box.value" assembling {
 search hole "offset" grammar "integer-hole-residual/v1" intent "Add one before doubling." max_candidates "8"
 case "2" -> "6" attempts "8"
}
activity Main(Integer) -> Integer computes "return Build(input)"
`

func TestJointSearchLegacyPoolUsesOriginalPlanningRoute(t *testing.T) {
	ctx := context.Background()
	source := []byte(searchValidationSource)
	cases := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":10}}]`)
	initial, err := GenerateCompositionWithOptions(ctx, "legacy.gooo", source, cases, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	search := initial.Preparations[0].Generation.Report.BodySearch
	search.CandidateSet = nil
	slots, _, err := jointSlots(ctx, "legacy.gooo", source, initial)
	if err != nil || len(slots) != 1 || slots[0].searchReceipt != nil {
		t.Fatal("historical missing pool acquired a new selection mode", slots, err)
	}
	expected, _, err := bodycodegen.RealizeSourceSearchCandidate(ctx, "legacy.gooo", source, "Build", search.SelectedCandidateID)
	if err != nil {
		t.Fatal(err)
	}
	actual, _, err := realizeJointSearch(ctx, "legacy.gooo", source, slots[0], search.SelectedCandidateID)
	if err != nil || actual.PlanningSourceSHA256 != "" || search.CandidateSet != nil ||
		!bodycodegen.EqualSearchCandidateHistory([]bodycodegen.SearchCandidate{expected}, []bodycodegen.SearchCandidate{actual}) {
		t.Fatal("historical planning or captured probes changed", actual, err)
	}
}

func TestJointSearchPlanningUsesSelectedHelperCheckpoint(t *testing.T) {
	source := []byte(jointSearchCheckpointSource)
	cases := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":10}}]`)
	r, err := ConstructJointComposition(context.Background(), "checkpoint.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 8, GoBinary: nativeTool()})
	if err != nil || r.Decision != "COMPLETE_FINITE" || r.SelectedAttempt != 0 {
		t.Fatal("selected helper source was lost while planning caller candidates", r.Stage, r.Failure, err)
	}
	initial := r.Initial.Preparations[1]
	search := initial.Generation.Report.BodySearch
	if search.CandidateSet == nil || search.OriginalSourceDigest == digest(source) ||
		search.OriginalSourceDigest != initial.InputSourceSHA256 || len(r.Attempts[0].SearchCandidates) != 1 {
		t.Fatal("candidate set lost its prepared helper checkpoint", search)
	}
	chosen := r.Attempts[0].SearchCandidates[0]
	if chosen.PlanningSourceSHA256 != initial.InputSourceSHA256 || chosen.PlanSHA256 != search.IRPlanSHA256 ||
		chosen.LocalCalls == nil || len(chosen.LocalCalls.Evaluations) != 1 ||
		chosen.LocalCalls.Evaluations[0].Phase != "candidate" {
		t.Fatal("caller selection regenerated the pool or lost its origin", chosen)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayJointComposition(context.Background(), "checkpoint.gooo", source, saved, cases, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 1 {
		t.Fatal("checkpoint candidate set did not replay", replayed, err)
	}
	for name, change := range map[string]func(*JointConstruction){
		"pool": func(c *JointConstruction) {
			c.Initial.Preparations[1].Generation.Report.BodySearch.CandidateSet.Candidates[0].Expression = "99"
		},
		"origin": func(c *JointConstruction) { c.Attempts[0].SearchCandidates[0].PlanningSourceSHA256 = "changed" },
		"calls": func(c *JointConstruction) {
			c.Attempts[0].SearchCandidates[0].LocalCalls.Evaluations[0].Calls[0].Inputs = json.RawMessage("[99]")
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&changed)
			if _, err := ReplayJointComposition(context.Background(), "checkpoint.gooo", source, changed, cases, nativeTool()); err == nil {
				t.Fatal("modified candidate history was accepted")
			}
		})
	}
	// A later combination may change the helper. Keep the initial bounded pool,
	// but evaluate the expression against that combination's actual helper body.
	slots, _, err := jointSlots(context.Background(), "checkpoint.gooo", source, r.Initial)
	if err != nil {
		t.Fatal(err)
	}
	attempt, _, _, err := materializeJoint(context.Background(), "checkpoint.gooo", source, cases, "Main", slots,
		[]uint16{0, slots[1].initial})
	if err != nil || attempt.Rejection != nil || len(attempt.SearchCandidates) != 1 {
		t.Fatal("changed helper could not use the retained pool", attempt, err)
	}
	changed := attempt.SearchCandidates[0]
	if changed.InputSourceSHA256 == chosen.InputSourceSHA256 || changed.PlanningSourceSHA256 != chosen.PlanningSourceSHA256 ||
		changed.PlanSHA256 != chosen.PlanSHA256 || changed.Attempt.TestCasesPassed != 0 || changed.Attempt.CaseResults[0].Actual != 3 {
		t.Fatal("changed helper reused an old outcome or changed the pool", changed)
	}
}
