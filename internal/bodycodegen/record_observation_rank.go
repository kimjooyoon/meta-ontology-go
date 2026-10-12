package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type recordObservationRanker struct {
	plan       recordAssemblyPlan
	receipt    *RecordObservationReceipt
	outputs    map[string]RecordProbeOutput
	candidates map[uint16]generatedRoute
}

// Candidate output is only a partition of the permitted hypotheses. It never
// becomes an expected value; the separate fixed oracle owns acquired answers.
func (e *recordObservationRanker) rank(ctx context.Context, cases []assemblyspec.ValueCase) (RecordObservationRound, error) {
	count := min(e.receipt.Options.MaxCandidates, 1<<len(e.plan.choices))
	r := RecordObservationRound{Candidates: count, Unobserved: (1 << len(e.plan.choices)) - count}
	for mask := range count {
		matched, unavailable := true, false
		for _, c := range cases {
			value, err := e.output(ctx, uint16(mask), c.Inputs)
			if err != nil {
				return r, err
			}
			if value.Error != "" {
				r.Unavailable = append(r.Unavailable, value)
				unavailable = true
				break
			}
			expected, err := assemblyspec.RecordValueKey(c.Expected)
			if err != nil {
				return r, err
			}
			if value.Value != expected {
				matched = false
				break
			}
		}
		if unavailable {
			r.UnavailableMasks = append(r.UnavailableMasks, uint16(mask))
		} else if matched {
			r.SurvivingMasks = append(r.SurvivingMasks, uint16(mask))
		}
	}
	used := make(map[string]bool, len(cases))
	for _, c := range cases {
		key, _ := assemblyspec.RecordValueKey(c.Inputs)
		used[key] = true
	}
	best := 0
	for index, input := range e.receipt.Options.ValueInputs {
		key, _ := assemblyspec.RecordValueKey(input)
		probe := RecordObservationProbe{Input: input, AlreadyTested: used[key]}
		if !probe.AlreadyTested {
			groups, available := map[string]int{}, 0
			for _, mask := range r.SurvivingMasks {
				value, err := e.output(ctx, mask, input)
				if err != nil {
					return r, err
				}
				probe.Outputs = append(probe.Outputs, value)
				if value.Error == "" {
					probe.SeparatedPairs += available - groups[value.Value]
					groups[value.Value]++
					available++
				}
			}
		}
		r.Probes = append(r.Probes, probe)
		if probe.SeparatedPairs > best {
			best = probe.SeparatedPairs
			r.RecommendedIndex = new(index)
		}
	}
	return r, ctx.Err()
}

func (e *recordObservationRanker) output(ctx context.Context, mask uint16, input string) (RecordProbeOutput, error) {
	if err := ctx.Err(); err != nil {
		return RecordProbeOutput{}, err
	}
	key, err := assemblyspec.RecordValueKey(input)
	if err != nil {
		return RecordProbeOutput{}, err
	}
	key = fmt.Sprintf("%d:%s", mask, key)
	if value, ok := e.outputs[key]; ok && e.receipt.Options.ReuseProbeOutputs {
		e.receipt.ReusedOutputs++
		return value, nil
	}
	value := RecordProbeOutput{Mask: mask}
	candidate, ok := e.candidates[mask]
	if !ok {
		candidate, err = e.plan.candidate(mask)
		if err == nil {
			e.candidates[mask] = candidate
		}
	}
	if err == nil {
		e.receipt.CandidateEvaluations++
		var cases []RecordAssemblyCase
		cases, err = evaluateRecordAssemblyRecorded(ctx, candidate.source, candidate.report.Activity, candidate.report.RecordTypes,
			[]assemblyspec.ValueCase{{Inputs: input, Expected: e.plan.spec.ValueCases[0].Expected}}, nil)
		if err == nil {
			value.Value, err = assemblyspec.RecordValueKey(string(cases[0].Actual))
		}
	}
	if err != nil {
		if ctx.Err() != nil {
			return value, ctx.Err()
		}
		value.Error = err.Error()
	}
	if e.receipt.Options.ReuseProbeOutputs {
		e.outputs[key] = value
	}
	return value, nil
}
