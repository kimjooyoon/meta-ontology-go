package bodyexecution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// ConstructJointComposition uses real caller executions to reconsider locally
// satisfactory helpers. Each completed combination is executed immediately.
// Optional inference belongs to Initial only; feedback controls bounded search.
func ConstructJointComposition(ctx context.Context, filename string, source []byte,
	cases CompositionCases, options JointOptions) (JointConstruction, error) {
	started := time.Now()
	r := JointConstruction{Schema: jointSchema, Stage: "PREFLIGHT", OriginalSourceSHA256: digest(source),
		ConstructionCases: cases, ConstructionSHA256: compositionDigest(cases),
		ProgramBudget: options.ProgramBudget, SelectedAttempt: -1,
		Scope: "source-bounded record combinations; local obligations and native caller expectations retained separately; initial optional model order; caller cases are consumed construction feedback; no general correctness claim"}
	finish := func(err error) (JointConstruction, error) {
		r.ElapsedNS = time.Since(started).Nanoseconds()
		if err != nil {
			r.Failure = err.Error()
		}
		return r, err
	}
	if ctx == nil || options.ProgramBudget < 1 || options.ProgramBudget > 64 {
		return finish(fmt.Errorf("joint construction requires context and 1..64 program attempts"))
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := validateJointSource(ctx, filename, source, cases, options.EntryActivity); err != nil {
		return finish(err)
	}
	var err error
	r.Stage = "INITIAL_LOCAL_CONSTRUCTION"
	r.Initial, err = GenerateCompositionWithOptions(ctx, filename, source, cases,
		CompositionOptions{EntryActivity: options.EntryActivity, ModelPath: options.ModelPath, FillModelPath: options.FillModelPath})
	if err != nil {
		return finish(err)
	}
	slots, space, err := jointSlots(ctx, filename, source, r.Initial)
	if err != nil {
		return finish(err)
	}
	r.CandidateSpace, r.Stage = space, "CALLER_GUIDED_SEARCH"
	r.CandidateKinds = jointCandidateKinds(slots)
	r.Schema = jointObservationSchema(slots, nil)
	if len(r.CandidateKinds) != 0 {
		r.Scope = "source-bounded record masks and integer IR search expressions; local obligations and native caller expectations retained separately; optional model order only for record choices; search uses its deterministic source-budget prefix; no general correctness claim"
	}
	if r.Schema == jointFillSchema || r.Schema == jointFillRejectionSchema {
		r.Scope = "source-bounded fill assignments, record masks and integer search expressions; selection uses local training and native caller cases only; local fill holdout observations do not influence selection or completion; optional models select initial bodies; subsequent fills use source order; rejected search rows have no caller score and local totals cover their scored prefix; finite observations only"
	}
	if r.Schema == jointFillRejectionSchema {
		r.Scope += "; rejected fill assignments consume attempts without native execution or a local score; all source assignments remain in the candidate space"
	}
	if r.Schema == jointPathSchema {
		r.Scope = "source-bounded typed paths, record masks, search expressions and fill assignments; initial local model observations retained unchanged; typed paths use the initial selected mask first, then deterministic fallback distance and numeric order within the source attempt cap; record ranking and source search/fill orders retain their own contracts; local cases and native caller cases remain separate; no model calls in subsequent candidates or saved replay; rejected combined paths consume attempts without caller scores; finite observations only"
	}
	executor := NewExecutor()
	defer executor.Close()
	for _, masks := range jointMaskOrder(slots, options.ProgramBudget) {
		if err := ctx.Err(); err != nil {
			return finish(err)
		}
		attempt, selected, program, err := materializeJoint(ctx, filename, source, cases, options.EntryActivity, slots, masks)
		if err != nil {
			return finish(err)
		}
		if attempt.Rejection == nil {
			attempt.Runtime, err = executor.ExecuteComposition(ctx, filename, selected, program, cases, options.GoBinary)
		} else {
			if r.Schema != jointRejectionSchema && r.Schema != jointFillSchema && r.Schema != jointFillRejectionSchema {
				r.Scope += "; rejected local expressions consume attempts without a caller score; rejected local totals cover only the scored prefix"
			}
		}
		r.Attempts = append(r.Attempts, attempt)
		if r.Schema != jointFaultSchema && hasCompositionFault(attempt.Runtime) {
			r.Scope += "; native integer zero-divisor outcomes preserve reached operation identities and operands; dependent activities are blocked; twice-reproduced language faults consume program attempts and permit continuation"
		}
		r.Schema = jointObservationSchema(slots, r.Attempts)
		if err != nil {
			return finish(err)
		}
		if attempt.Rejection == nil && (r.SelectedAttempt < 0 || betterJoint(attempt, r.Attempts[r.SelectedAttempt])) {
			r.SelectedAttempt, r.SelectedSource, r.Selected = len(r.Attempts)-1, string(selected), program
		}
		if raw, err := json.MarshalIndent(r, "", "  "); err != nil || len(raw) > 30<<20 {
			return finish(fmt.Errorf("joint construction exceeds the 32 MiB observation limit"))
		}
		if jointComplete(attempt) {
			break
		}
	}
	r.Decision, r.StopReason = jointOutcome(r)
	r.Stage = "COMPLETE"
	return finish(ctx.Err())
}

func validateJointSource(ctx context.Context, filename string, source []byte, cases CompositionCases, entry string) error {
	if cases.Schema != "gooo/body-composition-cases/v1" {
		return fmt.Errorf("joint construction needs explicit caller expectations")
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, entry)
	if err != nil {
		return err
	}
	if _, err = graph.inputRows(cases); err != nil {
		return err
	}
	var names []string
	for _, helper := range graph.plan.Preparations {
		names = append(names, helper.Name)
	}
	for _, node := range graph.nodes[:graph.count] {
		if node.Assembling && !node.Prepared {
			names = append(names, node.Name)
		}
	}
	if len(names) == 0 || len(names) > compositionLimit {
		return fmt.Errorf("joint construction requires 1..16 source assembly bodies")
	}
	for _, name := range names {
		spec, err := bodycodegen.SourceAssembly(ctx, filename, source, name)
		if err != nil || !bodycodegen.IsRecordAssembly(spec) && !bodycodegen.IsSourceIRSearch(spec) && !bodycodegen.IsSourceIRBodyFill(spec) && !bodycodegen.IsSourceTypedPathAssembly(spec) {
			return fmt.Errorf("joint construction requires a source assembly contract at %s", name)
		}
	}
	return nil
}

func materializeJoint(ctx context.Context, filename string, source []byte, cases CompositionCases,
	entry string, slots []jointSlot, masks []uint16) (JointAttempt, []byte, Composition, error) {
	attempt := JointAttempt{Masks: append([]uint16(nil), masks...)}
	if len(masks) != len(slots) {
		return attempt, nil, Composition{}, fmt.Errorf("joint candidate arity differs")
	}
	current := source
	for i, slot := range slots {
		if len(slot.pathMasks) != 0 {
			if !slices.Contains(slot.pathMasks, masks[i]) {
				return attempt, nil, Composition{}, fmt.Errorf("typed path mask exceeds the source attempt cap")
			}
			candidate, selected, err := bodycodegen.RealizeSourcePathCandidate(ctx, filename, current, slot.activity, slot.pathReceipt, masks[i])
			attempt.PathCandidates = append(attempt.PathCandidates, candidate)
			if err != nil {
				if rejected, ok := errors.AsType[*bodycodegen.SourcePathCandidateRejection](err); ok && ctx.Err() == nil {
					attempt.Rejection = &JointCandidateRejection{Stage: "LOCAL_TYPED_PATH", Slot: i,
						Activity: slot.activity, CandidateID: fmt.Sprintf("path_%04x", masks[i]), Reason: rejected.Error()}
					return attempt, nil, Composition{}, nil
				}
				return attempt, nil, Composition{}, fmt.Errorf("joint typed path activity %s: %w", slot.activity, err)
			}
			attempt.LocalPassed += candidate.Passed
			attempt.LocalTotal += candidate.Total
			current = selected
			continue
		}
		if len(slot.fillIDs) != 0 {
			if int(masks[i]) >= len(slot.fillIDs) {
				return attempt, nil, Composition{}, fmt.Errorf("fill candidate index exceeds the source bound")
			}
			candidate, selected, err := bodycodegen.RealizeSourceFillCandidate(ctx, filename, current, slot.activity, slot.fillIDs[masks[i]])
			if err != nil {
				var rejected *bodycodegen.SourceFillCandidateRejection
				if errors.As(err, &rejected) && ctx.Err() == nil {
					attempt.FillCandidates = append(attempt.FillCandidates, candidate)
					attempt.Rejection = &JointCandidateRejection{Stage: "LOCAL_SOURCE_FILL", Slot: i,
						Activity: slot.activity, CandidateID: candidate.CandidateID, Reason: rejected.Error()}
					return attempt, nil, Composition{}, nil
				}
				return attempt, nil, Composition{}, fmt.Errorf("joint fill activity %s: %w", slot.activity, err)
			}
			attempt.FillCandidates = append(attempt.FillCandidates, candidate)
			attempt.LocalPassed += candidate.TestCasesPassed
			attempt.LocalTotal += candidate.TestCasesTotal
			current = selected
			continue
		}
		if len(slot.searchIDs) != 0 {
			if int(masks[i]) >= len(slot.searchIDs) {
				return attempt, nil, Composition{}, fmt.Errorf("search candidate index exceeds the source bound")
			}
			candidate, selected, err := bodycodegen.RealizeSourceSearchCandidate(ctx, filename, current, slot.activity, slot.searchIDs[masks[i]])
			if err != nil {
				var rejected *bodycodegen.SourceSearchCandidateRejection
				if errors.As(err, &rejected) && ctx.Err() == nil {
					attempt.SearchCandidates = append(attempt.SearchCandidates, candidate)
					attempt.Rejection = &JointCandidateRejection{Stage: "LOCAL_SOURCE_SEARCH", Slot: i,
						Activity: slot.activity, CandidateID: candidate.Attempt.CandidateID, Reason: rejected.Error()}
					return attempt, nil, Composition{}, nil
				}
				return attempt, nil, Composition{}, fmt.Errorf("joint search activity %s: %w", slot.activity, err)
			}
			attempt.SearchCandidates = append(attempt.SearchCandidates, candidate)
			attempt.LocalPassed += candidate.Attempt.TestCasesPassed
			attempt.LocalTotal += candidate.Attempt.TestCasesTotal
			current = selected
			continue
		}
		candidate, selected, err := bodycodegen.RealizeRecordCandidate(ctx, filename, current, slot.activity, masks[i])
		if err != nil {
			return attempt, nil, Composition{}, fmt.Errorf("joint activity %s: %w", slot.activity, err)
		}
		appendJointLocal(&attempt, candidate)
		current = selected
	}
	program, err := GenerateCompositionWithOptions(ctx, filename, current, cases, CompositionOptions{EntryActivity: entry})
	return attempt, current, program, err
}

func jointOutcome(r JointConstruction) (string, string) {
	if r.SelectedAttempt >= 0 && jointComplete(r.Attempts[r.SelectedAttempt]) {
		return "COMPLETE_FINITE", "LOCAL_AND_CALLER_CASES_MATCHED"
	}
	if r.CandidateSpace == strconv.Itoa(len(r.Attempts)) {
		return "PARTIAL_FINITE", "DECLARED_SPACE_EXHAUSTED"
	}
	return "PARTIAL_FINITE", "PROGRAM_BUDGET_EXHAUSTED"
}
