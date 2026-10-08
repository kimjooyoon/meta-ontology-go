package bodyexecution

import (
	"context"
	"fmt"
	"math/big"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type jointSlot struct {
	activity  string
	initial   uint16
	ranking   []uint16
	searchIDs []string
}

func jointSlots(ctx context.Context, filename string, source []byte, prior Composition) ([]jointSlot, string, error) {
	var slots []jointSlot
	space := big.NewInt(1)
	for _, step := range prior.ConstructionSteps() {
		if search := step.Generation.Report.BodySearch; search != nil {
			set, err := bodycodegen.PlanSourceSearchCandidates(ctx, filename, source, step.Generation.Report.Activity)
			if err != nil {
				return nil, "", err
			}
			if search.IRPlanSHA256 != set.PlanSHA256 || search.AttemptBudget == nil || *search.AttemptBudget != set.AttemptBudget || search.CandidateCount != len(set.Candidates) {
				return nil, "", fmt.Errorf("joint source search contract differs from initial construction")
			}
			slot := jointSlot{activity: step.Generation.Report.Activity}
			found := false
			for i, candidate := range set.Candidates[:min(set.AttemptBudget, len(set.Candidates))] {
				slot.searchIDs = append(slot.searchIDs, candidate.ID)
				slot.ranking = append(slot.ranking, uint16(i))
				if candidate.ID == search.SelectedCandidateID {
					slot.initial, found = uint16(i), true
				}
			}
			if !found {
				return nil, "", fmt.Errorf("initial search selection exceeds the source budget")
			}
			slots = append(slots, slot)
			space.Mul(space, big.NewInt(int64(len(slot.ranking))))
			continue
		}
		r := step.Generation.Report.RecordAssembly
		if r == nil {
			if step.Generation.Report.BodyFill != nil || step.Generation.Report.BodyPaths != nil {
				return nil, "", fmt.Errorf("joint construction requires record-choice or source IR search bodies")
			}
			continue
		}
		if r.AttemptBudget == nil || *r.AttemptBudget < 1 || len(r.Ranking) == 0 {
			return nil, "", fmt.Errorf("joint construction needs an explicit source-bound choice budget")
		}
		ranking := slices.Clone(r.Ranking[:min(*r.AttemptBudget, len(r.Ranking))])
		if !slices.Contains(ranking, r.SelectedMask) {
			return nil, "", fmt.Errorf("initial selection exceeds the source budget")
		}
		slots = append(slots, jointSlot{activity: step.Generation.Report.Activity, initial: r.SelectedMask, ranking: ranking})
		space.Mul(space, big.NewInt(int64(len(ranking))))
	}
	if len(slots) < 1 || len(slots) > compositionLimit {
		return nil, "", fmt.Errorf("joint construction requires 1..16 record-choice or source IR search bodies")
	}
	return slots, space.String(), nil
}

func jointCandidateKinds(slots []jointSlot) []string {
	var kinds []string
	hasSearch := false
	for _, slot := range slots {
		kind := "record_mask"
		if len(slot.searchIDs) != 0 {
			kind, hasSearch = "source_search_index", true
		}
		kinds = append(kinds, kind)
	}
	if !hasSearch {
		return nil
	}
	return kinds
}

// Enumerate bounded prefixes directly, without allocating the Cartesian space.
// The locally chosen program is first; remaining tuples follow the original
// per-activity model order (or deterministic order), with no duplicate first row.
func jointMaskOrder(slots []jointSlot, budget int) [][]uint16 {
	first := make([]uint16, len(slots))
	for i, slot := range slots {
		first[i] = slot.initial
	}
	result := [][]uint16{first}
	indices := make([]int, len(slots))
	for len(result) < budget {
		row := make([]uint16, len(slots))
		for i, slot := range slots {
			row[i] = slot.ranking[indices[i]]
		}
		if !slices.Equal(row, first) {
			result = append(result, row)
		}
		advanced := false
		for i := len(indices) - 1; i >= 0; i-- {
			indices[i]++
			if indices[i] < len(slots[i].ranking) {
				advanced = true
				break
			}
			indices[i] = 0
		}
		if !advanced {
			break
		}
	}
	return result
}

func jointLocalComplete(attempt JointAttempt) bool {
	return attempt.LocalTotal > 0 && attempt.LocalPassed == attempt.LocalTotal
}

func jointComplete(attempt JointAttempt) bool {
	return jointLocalComplete(attempt) && attempt.Runtime.FiniteTotal > 0 &&
		attempt.Runtime.FinitePassed == attempt.Runtime.FiniteTotal
}

func betterJoint(a, b JointAttempt) bool {
	if jointLocalComplete(a) != jointLocalComplete(b) {
		return jointLocalComplete(a)
	}
	if a.Runtime.FinitePassed != b.Runtime.FinitePassed {
		return a.Runtime.FinitePassed > b.Runtime.FinitePassed
	}
	return a.LocalPassed > b.LocalPassed
}

func appendJointLocal(attempt *JointAttempt, candidate bodycodegen.RecordCandidate) {
	attempt.Candidates = append(attempt.Candidates, candidate)
	attempt.LocalPassed += candidate.Attempt.Passed
	attempt.LocalTotal += candidate.Attempt.Total
}
