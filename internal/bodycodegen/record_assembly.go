package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

// GenerateSourceAssembly dispatches the source-owned scalar or record contract.
// The retained immutable model is shared; all selection state belongs to a call.
func (g *TypedPathGenerator) GenerateSourceAssembly(ctx context.Context, filename string, source []byte, activity string) (Result, error) {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return Result{}, err
	}
	if IsSourceIRBodyFill(spec) {
		return Result{}, fmt.Errorf("source_fill requires GenerateWithSourceIRBodyFill; typed path models rank scalar or record choices")
	}
	if IsSourceIRSearch(spec) {
		return Result{}, fmt.Errorf("search requires GenerateWithSourceIRSearch; typed path models rank scalar or record choices")
	}
	if IsRecordAssembly(spec) {
		return g.generateRecordAssembly(ctx, filename, source, activity)
	}
	document, err := DecodeSourcePathDocument(ctx, filename, source, activity, nil)
	if err != nil {
		return Result{}, err
	}
	return g.Generate(ctx, filename, source, activity, document, TypedPathOptions{})
}

func (g *TypedPathGenerator) generateRecordAssembly(ctx context.Context, filename string, source []byte, activity string) (Result, error) {
	return g.generateRecordAssemblyWithPolicy(ctx, filename, source, activity, nil)
}

// GenerateRecordAssemblyWithPolicy evaluates the explicit Gooo policy after each
// scored candidate, before constructing another. The source budget stays fixed.
func (g *TypedPathGenerator) GenerateRecordAssemblyWithPolicy(ctx context.Context, filename string, source []byte,
	activity string, policy RecordAssemblyPolicy) (Result, error) {
	return g.generateRecordAssemblyWithPolicy(ctx, filename, source, activity, &policy)
}

func (g *TypedPathGenerator) generateRecordAssemblyWithPolicy(ctx context.Context, filename string, source []byte,
	activity string, policy *RecordAssemblyPolicy) (Result, error) {
	started := time.Now()
	if g == nil || g.info.Schema != "gooo/retained-path-model/v1" {
		return Result{}, fmt.Errorf("record assembly generator required")
	}
	if ctx == nil {
		return Result{}, fmt.Errorf("record assembly context required")
	}
	ctx, cancel := context.WithTimeout(ctx, bodyPathBudget)
	defer cancel()
	plan, err := prepareRecordAssembly(ctx, filename, source, activity)
	if err != nil {
		return Result{}, err
	}
	r := newRecordAssemblyReceipt(source, plan)
	var controller *preparedRecordPolicy
	if policy != nil {
		controller, err = prepareRecordPolicy(ctx, *policy)
		if err != nil {
			return Result{}, err
		}
		r.Control = controller.receipt
	}
	if err = observeRecordAssembly(ctx, filename, source, &plan, r); err != nil {
		return Result{}, &RecordAssemblyError{Receipt: r, Cause: err}
	}
	if err = g.rankRecordAssemblyContext(ctx, plan, r); err != nil {
		return Result{}, &RecordAssemblyError{Receipt: r, Cause: err}
	}
	if err = searchRecordAssemblyWithPolicy(ctx, plan, r, controller); err != nil {
		return Result{}, &RecordAssemblyError{Receipt: r, Cause: err}
	}
	result, err := emitSelectedRecordAssembly(ctx, filename, source, plan, r)
	r.GenerationNS = time.Since(started).Nanoseconds()
	if err == nil {
		populateCompletenessReceipt(&result.Report, "")
	}
	return result, err
}

func (g *TypedPathGenerator) rankRecordAssembly(plan recordAssemblyPlan, r *RecordAssemblyReceipt) error {
	return g.rankRecordAssemblyContext(context.Background(), plan, r)
}

func (g *TypedPathGenerator) rankRecordAssemblyContext(ctx context.Context, plan recordAssemblyPlan, r *RecordAssemblyReceipt) error {
	r.ModelRequested = g.info.Loaded
	if resolveRecordObservationRanking(r) {
		if g.info.Loaded {
			info := g.Info()
			r.Model = &info
		}
		return nil
	}
	if g.criteria != nil {
		return g.rankScalarCriteria(ctx, plan, r)
	}
	if g.info.Loaded {
		info := g.Info()
		r.Model = &info
		r.Context = recordPlanModelContext(plan, info.FeatureVersion)
		if g.three == nil && plan.scalar == nil {
			r.Context.Status, r.Context.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_CHOICE_MODEL_REQUIRED"
		}
		if r.Context.Status == "ENCODED" {
			var workspace jointdecision.ThreeWorkspace
			var prediction jointdecision.ThreePrediction
			var err error
			predictStarted := time.Now()
			switch info.FeatureVersion {
			case jointdecision.RecordFieldFeatureVersion:
				err = g.three.PredictRecordInto(r.Context.Text, &workspace, &prediction)
			case jointdecision.RecordSharedFeatureVersion:
				err = g.three.PredictRecordSharedInto(r.Context.Text, &workspace, &prediction)
			case jointdecision.RecordOriginSharedFeatureVersion:
				err = g.three.PredictRecordOriginSharedInto(r.Context.Text, &workspace, &prediction)
			case jointdecision.RecordGraphSharedFeatureVersion:
				err = g.three.PredictRecordGraphSharedInto(r.Context.Text, &workspace, &prediction)
			default:
				err = g.three.PredictInto(r.Context.Text, &workspace, &prediction)
			}
			r.ModelCalls, r.PredictNS = 1, time.Since(predictStarted).Nanoseconds()
			if err != nil {
				return fmt.Errorf("record field prediction: %w", err)
			}
			r.Prediction = &prediction
			slices.SortFunc(r.Ranking, func(a, b uint16) int {
				if prediction.Probabilities[a] > prediction.Probabilities[b] {
					return -1
				}
				if prediction.Probabilities[a] < prediction.Probabilities[b] {
					return 1
				}
				return int(a) - int(b)
			})
		}
	}
	return nil
}

func newRecordAssemblyReceipt(source []byte, p recordAssemblyPlan) *RecordAssemblyReceipt {
	contract, _ := p.spec.Canonical()
	cases, _ := json.Marshal(criteriaValueCases(p.spec, false))
	budget := p.spec.MaxAttempts
	r := &RecordAssemblyReceipt{Schema: recordAssemblySchema, OriginalSourceSHA256: digest(source),
		ContractSHA256: digest([]byte(contract)), TestSuiteSHA256: digest(cases), AttemptBudget: &budget,
		Choices: append([]RecordValueChoice(nil), p.choices...),
		Scope:   "source-declared record field alternatives; finite case and field coverage; candidate interpretation followed by ordinary typed Go emission; no model updates"}
	if p.scalar != nil {
		r.Schema = scalarCriteriaAssemblySchema
		r.Scope = "source-declared Integer path choices with actual scalar results and fixed Gooo predicates; finite source inputs; record_assembly is the legacy receipt container; no record fields or model predicate understanding are implied"
	}
	for mask := 0; mask < 1<<len(p.choices); mask++ {
		r.Ranking = append(r.Ranking, uint16(mask))
	}
	return r
}

func searchRecordAssembly(ctx context.Context, p recordAssemblyPlan, r *RecordAssemblyReceipt) error {
	return searchRecordAssemblyWithPolicy(ctx, p, r, nil)
}

func searchRecordAssemblyWithPolicy(ctx context.Context, p recordAssemblyPlan, r *RecordAssemblyReceipt, policy *preparedRecordPolicy) error {
	best := -1
	var policyBest RecordAssemblyAttempt
	var policyCases []RecordAssemblyCase
	remaining := r.Ranking[len(r.Attempts):]
	if len(r.Attempts) > 0 {
		var err error
		policyBest, policyCases, remaining, err = resumeRecordSearch(ctx, p, r, policy)
		if err != nil {
			return err
		}
		if usableRecordAttempt(policyBest) {
			best = 0
		}
	}
	for _, mask := range remaining {
		if len(r.Attempts) >= p.spec.MaxAttempts {
			break
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		attempt, cases, err := evaluateRecordMask(ctx, p, mask)
		if err != nil {
			if attempt.LocalCalls != nil {
				r.Attempts = append(r.Attempts, attempt)
			}
			return err
		}
		r.Attempts = append(r.Attempts, attempt)
		if !usableRecordAttempt(attempt) {
			continue
		}
		if policy != nil && (!usableRecordAttempt(policyBest) || betterRecordAttempt(attempt, policyBest, true)) {
			policyBest, policyCases = attempt, cases
		}
		if best < 0 || betterRecordAttempt(attempt, selectedRecordAttempt(r), false) {
			best = len(r.Attempts) - 1
			selectRecordAttempt(r, attempt, cases)
		}
		keepGoing := true
		if policy != nil {
			keepGoing, err = observeRecordPolicy(ctx, p, r, policy)
			if err != nil {
				return err
			}
			if r.Control.Decisions[len(r.Control.Decisions)-1].Operation == "USE_OBSERVED_CANDIDATE" {
				// The policy's best count refers to whole cases. Preserve that
				// meaning even when another candidate matches more individual fields.
				selectRecordAttempt(r, policyBest, policyCases)
			}
		}
		if !keepGoing || completeRecordAttempt(attempt) {
			break
		}
	}
	if best < 0 {
		return fmt.Errorf("record assembly has no valid typed candidate that completed evaluation within the attempt budget")
	}
	finishRecordAssemblySearch(r)
	return nil
}

func finishRecordAssemblySearch(r *RecordAssemblyReceipt) {
	r.Status = "PARTIAL_FINITE"
	if completeRecordAttempt(selectedRecordAttempt(r)) {
		r.Status = "COMPLETE_FINITE"
	}
	for i := range r.Choices {
		first, second := assemblyChoiceLabels(r.Choices[i].Kind)
		r.Choices[i].Picked = first
		if r.SelectedMask&(1<<i) != 0 {
			r.Choices[i].Picked = second
		}
	}
}

func scoreRecordCases(mask uint16, cases []RecordAssemblyCase) RecordAssemblyAttempt {
	r := RecordAssemblyAttempt{Mask: mask, Total: len(cases)}
	for _, c := range cases {
		if c.Passed {
			r.Passed++
		}
		r.FieldsTotal += len(c.Fields)
		for _, field := range c.Fields {
			if field.Passed {
				r.FieldsPassed++
			}
		}
	}
	return r
}

func emitRecordAssembly(ctx context.Context, filename string, source []byte, p recordAssemblyPlan, r *RecordAssemblyReceipt) (Result, error) {
	body, err := p.candidateBody(r.SelectedMask)
	if err != nil {
		return Result{}, err
	}
	choices := make(map[string]string, len(r.Choices))
	for _, c := range r.Choices {
		choices[c.ID] = c.Picked
	}
	completed, err := selectedTypedPathSource(source, p.original.activity, body, choices, true)
	if err != nil {
		return Result{}, err
	}
	result, err := GenerateWithPlanner(ctx, filename, completed, p.body.activity.Name, "", "")
	if err != nil {
		return Result{}, err
	}
	r.ChoiceOrigins, err = p.choiceOrigins(r.SelectedMask, body)
	if err != nil {
		return Result{}, err
	}
	r.SelectedSourceSHA256 = digest(completed)
	result.GoooSource, result.Report.RecordAssembly = string(completed), r
	return result, nil
}
