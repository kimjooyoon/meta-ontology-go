package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const jointFillExampleRoot = "examples/caller-source-fill"

func smokeJointSourceFill(binary, work string, input BuildInput) error {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(input.Root, jointFillExampleRoot, name))
	}
	source, err := read("budget.gooo.fixture")
	if err != nil {
		return err
	}
	feedback, err := read("construction-cases.json")
	if err != nil {
		return err
	}
	evaluation, err := read("evaluation-cases.json")
	if err != nil {
		return err
	}
	selected := ""
	for _, mode := range []string{"partial", "construct", "replay"} {
		budget := 3
		if mode == "partial" {
			budget = 2
		}
		directory := filepath.Join(work, "caller-fill-"+mode)
		args := []string{"body-construct", "--source", filepath.Join(jointFillExampleRoot, "budget.gooo.fixture"),
			"--cases", filepath.Join(jointFillExampleRoot, "evaluation-cases.json")}
		if mode == "replay" {
			args = append(args, "--construction", filepath.Join(work, "caller-fill-construct", "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--construction-cases", filepath.Join(jointFillExampleRoot, "construction-cases.json"),
				"--attempts", strconv.Itoa(budget), "--out", directory)
		}
		raw, runErr := commandOutput(input.Root, nil, binary, args...)
		if len(raw) > 0 {
			if err := os.WriteFile(filepath.Join(input.OutputDir, input.Target.ID+"-caller-fill-"+mode+".json"), raw, 0o644); err != nil {
				return err
			}
		}
		if runErr != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_FILL: %w", runErr)
		}
		selected, err = validateJointFillSmoke(raw, source, feedback, evaluation, budget, mode == "replay", selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_FILL: %w", err)
		}
	}
	return nil
}

type jointSmokeFillCase struct {
	Inputs           []json.RawMessage
	Actual, Expected json.RawMessage
	Passed           *bool
}

type jointSmokeFillCandidate struct {
	Rejection        *jointSmokeFillRejection
	Schema, Activity string
	ActivityID       string               `json:"activity_id"`
	InputSHA         string               `json:"input_source_sha256"`
	SelectedSHA      string               `json:"selected_source_sha256"`
	PlanSHA          string               `json:"plan_sha256"`
	ID               string               `json:"candidate_id"`
	Method           string               `json:"selection_method"`
	Count            *int                 `json:"candidate_count"`
	Holes            []jointSmokeFillHole `json:"hole_fills"`
	Passed           *int                 `json:"test_cases_passed"`
	Total            *int                 `json:"test_cases_total"`
	Cases            []json.RawMessage    `json:"case_results"`
	Values           []jointSmokeFillCase `json:"value_case_results"`
	HoldoutPassed    *int                 `json:"holdout_cases_passed"`
	HoldoutTotal     *int                 `json:"holdout_cases_total"`
	Holdout          []json.RawMessage    `json:"holdout_case_results"`
	ValueHoldout     []jointSmokeFillCase `json:"value_holdout_results"`
}

type jointSmokeFillHole struct {
	ID         string `json:"hole_id"`
	Expression string
}

type jointSmokeFillRejection struct {
	ID            string `json:"candidate_id"`
	Stage, Reason string
	Holes         []jointSmokeFillHole `json:"hole_fills"`
}
