package bodyexecution

import (
	"context"
	"fmt"
	"slices"
	"time"
)

func DecodeJointConstruction(raw []byte) (JointConstruction, error) {
	var result JointConstruction
	err := decode(raw, &result, 32<<20)
	return result, err
}

// ReplayJointComposition reconstructs every source combination and re-executes
// the consumed caller suite before executing a supplied evaluation suite. It
// neither loads a model nor treats consumed caller feedback as unseen data.
func ReplayJointComposition(ctx context.Context, filename string, source []byte, prior JointConstruction,
	cases CompositionCases, goBinary string) (JointEvaluation, error) {
	var result JointEvaluation
	if ctx == nil {
		return result, fmt.Errorf("joint replay requires context")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := verifyJointHeader(ctx, filename, source, prior); err != nil {
		return result, err
	}
	slots, space, err := jointSlots(ctx, filename, source, prior.Initial)
	if err != nil || space != prior.CandidateSpace {
		return result, fmt.Errorf("joint candidate space differs: %v", err)
	}
	kinds := jointCandidateKinds(slots)
	schema := jointSchema
	if len(kinds) != 0 {
		schema = jointMixedSchema
	}
	if prior.Schema != schema || !slices.Equal(prior.CandidateKinds, kinds) {
		return result, fmt.Errorf("joint candidate kinds or schema differ")
	}
	order := jointMaskOrder(slots, prior.ProgramBudget)
	if len(prior.Attempts) == 0 || len(prior.Attempts) > len(order) {
		return result, fmt.Errorf("joint attempt count exceeds the bound")
	}
	executor := NewExecutor()
	defer executor.Close()
	if err := verifyJointAttempts(ctx, filename, source, prior, slots, order, executor, goBinary); err != nil {
		return result, err
	}
	result.ConstructionReplayed = true
	result.Runtime, err = executor.ExecuteComposition(ctx, filename, []byte(prior.SelectedSource), prior.Selected, cases, goBinary)
	if err == nil {
		result.InputSeparation, err = measureJointInputs(ctx, filename, []byte(prior.SelectedSource), prior, cases)
	}
	return result, err
}

func verifyJointHeader(ctx context.Context, filename string, source []byte, r JointConstruction) error {
	if r.Schema != jointSchema && r.Schema != jointMixedSchema || r.Stage != "COMPLETE" || r.Failure != "" ||
		r.OriginalSourceSHA256 != digest(source) || r.ConstructionSHA256 != compositionDigest(r.ConstructionCases) ||
		r.ProgramBudget < 1 || r.ProgramBudget > 64 || r.SelectedAttempt < 0 || r.SelectedAttempt >= len(r.Attempts) {
		return fmt.Errorf("joint construction identity, stage or budget differs")
	}
	if err := validateJointSource(ctx, filename, source, r.ConstructionCases, r.Initial.Plan.EntryActivity); err != nil {
		return err
	}
	return VerifyComposition(ctx, filename, source, r.Initial)
}

func verifyJointAttempts(ctx context.Context, filename string, source []byte, prior JointConstruction,
	slots []jointSlot, order [][]uint16, executor *Executor, goBinary string) error {
	best := -1
	var bestSource []byte
	var bestProgram Composition
	for i, recorded := range prior.Attempts {
		if !slices.Equal(recorded.Masks, order[i]) {
			return fmt.Errorf("joint attempt %d differs from bounded source order", i)
		}
		attempt, selected, program, err := materializeJoint(ctx, filename, source, prior.ConstructionCases,
			prior.Initial.Plan.EntryActivity, slots, recorded.Masks)
		if err != nil {
			return err
		}
		if compositionDigest(attempt.Candidates) != compositionDigest(recorded.Candidates) ||
			compositionDigest(attempt.SearchCandidates) != compositionDigest(recorded.SearchCandidates) ||
			attempt.LocalPassed != recorded.LocalPassed || attempt.LocalTotal != recorded.LocalTotal {
			return fmt.Errorf("joint attempt %d local obligations differ", i)
		}
		attempt.Runtime, err = executor.ExecuteComposition(ctx, filename, selected, program, prior.ConstructionCases, goBinary)
		if err != nil {
			return err
		}
		if !sameJointRuntime(attempt.Runtime, recorded.Runtime) {
			return fmt.Errorf("joint attempt %d native caller observations differ", i)
		}
		if best < 0 || betterJoint(attempt, prior.Attempts[best]) {
			best, bestSource, bestProgram = i, selected, program
		}
		if jointComplete(attempt) && i != len(prior.Attempts)-1 {
			return fmt.Errorf("joint construction continued after satisfying all obligations")
		}
	}
	if best != prior.SelectedAttempt || string(bestSource) != prior.SelectedSource ||
		bestProgram.Source != prior.Selected.Source || bestProgram.Driver != prior.Selected.Driver {
		return fmt.Errorf("joint selected program differs from observed candidates")
	}
	decision, stop := jointOutcome(prior)
	if decision != prior.Decision || stop != prior.StopReason ||
		(decision != "COMPLETE_FINITE" && len(prior.Attempts) != len(order)) {
		return fmt.Errorf("joint stopping observation differs")
	}
	return VerifyComposition(ctx, filename, bestSource, prior.Selected)
}

func sameJointRuntime(a, b CompositionRuntime) bool {
	semantic := func(r CompositionRuntime) any {
		return []any{r.Stage, r.Failure, r.OriginalSourceSHA256, r.SelectedSourceSHA256,
			r.TypedPlanSHA256, r.GeneratedSHA256, r.DriverSHA256, r.RuntimeSuiteSHA256,
			r.Traces, r.FinitePassed, r.FiniteTotal, r.ModelCalls, r.ProjectionReplayed, r.RuntimeReplayed,
			r.InputSeparation}
	}
	return compositionDigest(semantic(a)) == compositionDigest(semantic(b))
}

func measureJointInputs(ctx context.Context, filename string, source []byte, prior JointConstruction,
	cases CompositionCases) (JointInputSeparation, error) {
	r := JointInputSeparation{Scope: "unique caller root-input tuples compared with consumed construction cases; local-example and model-training exposure are not established by this metric"}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, prior.Initial.Plan.EntryActivity)
	if err != nil {
		return r, err
	}
	old, err := graph.inputRows(prior.ConstructionCases)
	if err != nil {
		return r, err
	}
	rows, err := graph.inputRows(cases)
	if err != nil {
		return r, err
	}
	consumed, seen := map[string]bool{}, map[string]bool{}
	for _, row := range old {
		consumed[compositionDigest(row)] = true
	}
	for _, row := range rows {
		key := compositionDigest(row)
		if seen[key] {
			r.DuplicateRows++
			continue
		}
		seen[key], r.UniqueInputs = true, r.UniqueInputs+1
		if consumed[key] {
			r.ConstructionInputs++
		} else {
			r.OtherInputs++
		}
	}
	return r, nil
}
