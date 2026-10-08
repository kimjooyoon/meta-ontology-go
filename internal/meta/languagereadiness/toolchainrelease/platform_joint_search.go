package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const jointSearchExampleRoot = "examples/caller-search-rejection"

func smokeJointSearchRejection(binary, work string, input BuildInput) error {
	feedback, err := os.ReadFile(filepath.Join(input.Root, jointSearchExampleRoot, "construction-cases.json"))
	if err != nil {
		return err
	}
	evaluation, err := os.ReadFile(filepath.Join(input.Root, jointSearchExampleRoot, "evaluation-cases.json"))
	if err != nil {
		return err
	}
	directory := filepath.Join(work, "caller-search-rejection")
	selected := ""
	for _, replay := range []bool{false, true} {
		args := []string{"body-construct", "--source", filepath.Join(jointSearchExampleRoot, "main.gooo.fixture"),
			"--cases", filepath.Join(jointSearchExampleRoot, "evaluation-cases.json")}
		mode := "construct"
		if replay {
			mode = "replay"
			args = append(args, "--construction", filepath.Join(directory, "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--construction-cases", filepath.Join(jointSearchExampleRoot, "construction-cases.json"),
				"--attempts", "5", "--out", directory)
		}
		raw, err := commandOutput(input.Root, nil, binary, args...)
		if len(raw) > 0 {
			name := input.Target.ID + "-caller-search-" + mode + ".json"
			if writeErr := os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0o644); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_SEARCH: %w", err)
		}
		selected, err = validateJointSearchSmoke(raw, feedback, evaluation, replay, selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_SEARCH: %w", err)
		}
	}
	return nil
}

type jointSmokeRejection struct {
	Stage, Activity, Reason string
	Slot                    *int
	CandidateID             string `json:"candidate_id"`
}

type jointSmokeSearchCandidate struct {
	Schema, Activity string
	InputSHA         string `json:"input_source_sha256"`
	SelectedSHA      string `json:"selected_source_sha256"`
	PlanSHA          string `json:"plan_sha256"`
	Attempt          struct {
		ID         string   `json:"candidate_id"`
		Expression string   `json:"expression"`
		Error      string   `json:"error"`
		Typed      *bool    `json:"typecheck_passed"`
		Scored     *bool    `json:"scoring_completed"`
		Passed     *int     `json:"test_cases_passed"`
		Total      *int     `json:"test_cases_total"`
		Accuracy   *float64 `json:"accuracy_percent"`
		Cases      []struct {
			Input, Actual, Expected json.RawMessage
			Passed                  *bool
		} `json:"case_results"`
	}
}
