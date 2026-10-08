package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const jointExampleRoot = "examples/caller-guided-construction"

func smokeJointConstruction(binary, work string, input BuildInput) error {
	feedback, err := os.ReadFile(filepath.Join(input.Root, jointExampleRoot, "construction-cases.json"))
	if err != nil {
		return err
	}
	evaluation, err := os.ReadFile(filepath.Join(input.Root, jointExampleRoot, "evaluation-cases.json"))
	if err != nil {
		return err
	}
	directory := filepath.Join(work, "caller-construction")
	selected := ""
	for _, replay := range []bool{false, true} {
		args := []string{"body-construct", "--source", filepath.Join(jointExampleRoot, "main.gooo.fixture"),
			"--cases", filepath.Join(jointExampleRoot, "evaluation-cases.json")}
		mode := "construct"
		if replay {
			mode = "replay"
			args = append(args, "--construction", filepath.Join(directory, "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--construction-cases", filepath.Join(jointExampleRoot, "construction-cases.json"),
				"--attempts", "2", "--out", directory)
		}
		raw, err := commandOutput(input.Root, nil, binary, args...)
		if len(raw) > 0 {
			path := filepath.Join(input.OutputDir, input.Target.ID+"-caller-"+mode+".json")
			if writeErr := os.WriteFile(path, raw, 0o644); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_CONSTRUCTION: %w", err)
		}
		selected, err = validateJointSmoke(raw, feedback, evaluation, replay, selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_CONSTRUCTION: %w", err)
		}
	}
	return nil
}

type jointSmokeOutput struct {
	Generated    *bool `json:"generated_now"`
	Construction struct {
		Schema, Stage, Failure, Decision string
		StopReason                       string   `json:"stop_reason"`
		Budget                           *int     `json:"program_budget"`
		Space                            string   `json:"candidate_space"`
		Kinds                            []string `json:"candidate_kinds"`
		SelectedAttempt                  *int     `json:"selected_attempt"`
		Source                           string   `json:"selected_source"`
		Selected                         struct {
			SHA string `json:"generated_sha256"`
		}
		Initial  struct{ Model struct{ Loaded *bool } }
		Attempts []jointSmokeAttempt
	}
	Evaluation struct {
		Runtime    jointSmokeRuntime
		Replayed   *bool `json:"construction_replayed"`
		Calls      *int  `json:"new_model_calls"`
		Separation struct {
			Unique    *int `json:"unique_inputs"`
			Duplicate *int `json:"duplicate_rows"`
			Consumed  *int `json:"construction_inputs"`
			Other     *int `json:"other_inputs"`
		} `json:"input_separation"`
	}
}

type jointSmokeAttempt struct {
	Rejection        *jointSmokeRejection        `json:"rejection"`
	SearchCandidates []jointSmokeSearchCandidate `json:"search_candidates"`
	Masks            []int
	Passed           *int `json:"local_passed"`
	Total            *int `json:"local_total"`
	Candidates       []struct {
		Activity string
		Attempt  struct{ Mask, Passed, Total *int }
		Cases    []struct {
			Inputs           []json.RawMessage
			Actual, Expected json.RawMessage
			Passed           *bool
		}
	}
	Runtime jointSmokeRuntime
}

type jointSmokeRuntime struct {
	Stage, Failure string
	SHA            string `json:"generated_sha256"`
	Passed         *int   `json:"finite_passed"`
	Total          *int   `json:"finite_total"`
	Calls          *int   `json:"model_calls"`
	Projection     *bool  `json:"projection_replayed"`
	Replay         *bool  `json:"runtime_replayed"`
	Traces         []struct {
		Index      int `json:"case_index"`
		Deliveries []struct {
			ID                      string `json:"activity_id"`
			Input, Actual, Expected json.RawMessage
			Passed                  *bool
		}
	}
}
