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
	r.ModelRequested = g.info.Loaded
	if g.info.Loaded {
		info := g.Info()
		r.Model = &info
		r.Context = recordModelContext(plan.choices, info.FeatureVersion)
		if g.three == nil {
			r.Context.Status, r.Context.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_CHOICE_MODEL_REQUIRED"
		}
		if r.Context.Status == "ENCODED" {
			var workspace jointdecision.ThreeWorkspace
			var prediction jointdecision.ThreePrediction
			predictStarted := time.Now()
			if info.FeatureVersion == jointdecision.RecordFieldFeatureVersion {
				err = g.three.PredictRecordInto(r.Context.Text, &workspace, &prediction)
			} else {
				err = g.three.PredictInto(r.Context.Text, &workspace, &prediction)
			}
			r.ModelCalls, r.PredictNS = 1, time.Since(predictStarted).Nanoseconds()
			if err != nil {
				return Result{}, fmt.Errorf("record field prediction: %w", err)
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
	if err = searchRecordAssembly(ctx, plan, r); err != nil {
		return Result{}, err
	}
	result, err := emitRecordAssembly(ctx, filename, source, plan, r)
	r.GenerationNS = time.Since(started).Nanoseconds()
	if err == nil {
		populateCompletenessReceipt(&result.Report, "")
	}
	return result, err
}

func newRecordAssemblyReceipt(source []byte, p recordAssemblyPlan) *RecordAssemblyReceipt {
	contract, _ := p.spec.Canonical()
	cases, _ := json.Marshal(p.spec.ValueCases)
	r := &RecordAssemblyReceipt{Schema: recordAssemblySchema, OriginalSourceSHA256: digest(source),
		ContractSHA256: digest([]byte(contract)), TestSuiteSHA256: digest(cases), Choices: append([]RecordValueChoice(nil), p.choices...),
		Scope: "source-declared record field alternatives; finite case and field coverage; candidate interpretation followed by ordinary typed Go emission; no model updates"}
	for mask := 0; mask < 1<<len(p.choices); mask++ {
		r.Ranking = append(r.Ranking, uint16(mask))
	}
	return r
}

func searchRecordAssembly(ctx context.Context, p recordAssemblyPlan, r *RecordAssemblyReceipt) error {
	best := -1
	for _, mask := range r.Ranking {
		if len(r.Attempts) >= p.spec.MaxAttempts {
			break
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		candidate, err := p.candidate(mask)
		if err != nil {
			return err
		}
		cases, err := evaluateRecordAssembly(ctx, candidate.source, p.body.activity.Name, p.body.records, p.spec.ValueCases)
		if err != nil {
			return err
		}
		attempt := scoreRecordCases(mask, cases)
		r.Attempts = append(r.Attempts, attempt)
		if best < 0 || attempt.FieldsPassed > r.FieldsPassed || (attempt.FieldsPassed == r.FieldsPassed && attempt.Passed > r.Passed) {
			best = len(r.Attempts) - 1
			r.SelectedMask, r.Cases = mask, cases
			r.Passed, r.Total, r.FieldsPassed, r.FieldsTotal = attempt.Passed, attempt.Total, attempt.FieldsPassed, attempt.FieldsTotal
		}
		if attempt.Passed == attempt.Total {
			break
		}
	}
	if best < 0 {
		return fmt.Errorf("record assembly requires a candidate observation")
	}
	r.Status = "PARTIAL_FINITE"
	if r.Passed == r.Total {
		r.Status = "COMPLETE_FINITE"
	}
	for i := range r.Choices {
		r.Choices[i].Picked = "value_first"
		if r.SelectedMask&(1<<i) != 0 {
			r.Choices[i].Picked = "value_second"
		}
	}
	return nil
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
	r.SelectedSourceSHA256 = digest(completed)
	result.GoooSource, result.Report.RecordAssembly = string(completed), r
	return result, nil
}
