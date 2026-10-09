package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type planInputRuntime struct {
	bodyexecution.CompositionRuntime
	Passed    *int              `json:"finite_passed"`
	Total     *int              `json:"finite_total"`
	Calls     *int              `json:"model_calls"`
	Projected *bool             `json:"projection_replayed"`
	Replayed  *bool             `json:"runtime_replayed"`
	Processes []json.RawMessage `json:"runs"`
}

type planInputOutput struct {
	Generated   *bool                     `json:"generated_now"`
	InputSchema string                    `json:"input_schema"`
	Composition bodyexecution.Composition `json:"composition"`
	Runtime     planInputRuntime          `json:"runtime"`
	History     []planInputRuntime        `json:"runtime_history"`
}

func validatePlanInputExecution(raw []byte, plan bodyexecution.CompositionInspection,
	example planInputSmokeCase, mode, selected string) (string, error) {
	var result planInputOutput
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	c := result.Composition
	replay, scored := mode != "inputs", mode == "scored-replay"
	if result.Generated == nil || *result.Generated == replay || c.Stage != "COMPLETE" ||
		c.OriginalSourceSHA256 != plan.SourceSHA256 || !reflect.DeepEqual(c.Plan, plan.Plan) ||
		!nativeSmokeDigest(c.GeneratedSHA256) || selected != "" && selected != c.GeneratedSHA256 {
		return "", fmt.Errorf("input execution source, structural plan or saved program changed")
	}
	wantSchema := bodyexecution.CompositionInputsSchema
	if scored {
		wantSchema = "gooo/body-composition-cases/v1"
	}
	if result.InputSchema != wantSchema {
		return "", fmt.Errorf("input execution schema differs")
	}
	if err := validatePlanInputRuntime(result.Runtime, c, example.total, scored); err != nil {
		return "", err
	}
	if !replay {
		if err := validatePlanInputRepeated(result.History, c, example.total); err != nil {
			return "", err
		}
	}
	if example.entry == "Describe" {
		if err := validateScalarIdentityAssembly(c, example.model != ""); err != nil {
			return "", err
		}
	}
	return c.GeneratedSHA256, nil
}

func validatePlanInputRuntime(r planInputRuntime, c bodyexecution.Composition, total int, scored bool) error {
	want := 0
	if scored {
		want = total
	}
	if r.Stage != "COMPLETE" || r.OriginalSourceSHA256 != c.OriginalSourceSHA256 ||
		r.GeneratedSHA256 != c.GeneratedSHA256 || r.TypedPlanSHA256 != c.Plan.TypedPlanSHA256 ||
		!jointSmokeInt(r.Passed, want) || !jointSmokeInt(r.Total, want) || !jointSmokeInt(r.Calls, 0) ||
		!jointSmokeBool(r.Projected, true) || !jointSmokeBool(r.Replayed, true) {
		return fmt.Errorf("native input execution identity, counters or replay differ")
	}
	if err := validateNativeSmokeRuns(r.Processes); err != nil {
		return err
	}
	if !scored {
		return validatePlanUnscoredTraces(r, c.Plan)
	}
	return validatePlanScoredTraces(r, c.Plan, total)
}

func validatePlanInputRepeated(history []planInputRuntime, c bodyexecution.Composition, total int) error {
	if len(history) != 2 {
		return fmt.Errorf("input execution requires two retained runtime observations")
	}
	for _, r := range history {
		if err := validatePlanInputRuntime(r, c, total, false); err != nil {
			return err
		}
	}
	last := history[1]
	if last.Artifact == nil || !last.Artifact.Reused || last.Build.Started {
		return fmt.Errorf("repeated input execution did not reuse its retained executable")
	}
	return nil
}

func validatePlanUnscoredTraces(r planInputRuntime, plan bodyexecution.CompositionPlan) error {
	if r.InputSeparation.Status != "UNKNOWN" || r.InputSeparation.Reason != "NO_RUNTIME_EXPECTATIONS" ||
		len(r.Traces) != 2 || !planInputHasExactRoot(r.Traces[0]) {
		return fmt.Errorf("unscored input scope, rows or exact large integer changed")
	}
	for i, trace := range r.Traces {
		if trace.CaseIndex != i || len(trace.Deliveries) != len(plan.Activities) {
			return fmt.Errorf("unscored input delivery count or order differs")
		}
		for j, d := range trace.Deliveries {
			if d.ActivityID != plan.Activities[j].ID || !nativeSmokeNull(d.Expected) || d.Passed != nil ||
				d.Fault != nil || len(d.BlockedBy) != 0 || nativeSmokeNull(d.Actual) {
				return fmt.Errorf("unscored input acquired an oracle or lost a native result")
			}
		}
	}
	return nil
}

func planInputHasExactRoot(trace bodyexecution.CompositionTrace) bool {
	for _, delivery := range trace.Deliveries {
		if delivery.ProducerID == "" && string(delivery.Input) == "9007199254740993" {
			return true
		}
		for _, input := range delivery.Inputs {
			if input.ProducerID == "" && string(input.Value) == "9007199254740993" {
				return true
			}
		}
	}
	return false
}
