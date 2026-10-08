package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const jointFillRejectionRoot = "examples/caller-fill-rejection"

func smokeJointFillRejection(binary, work string, input BuildInput) error {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(input.Root, jointFillRejectionRoot, name))
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
		budget := 5
		if mode == "partial" {
			budget = 3
		}
		directory := filepath.Join(work, "caller-fill-rejection-"+mode)
		args := []string{"body-construct", "--source", filepath.Join(jointFillRejectionRoot, "budget.gooo.fixture"),
			"--cases", filepath.Join(jointFillRejectionRoot, "evaluation-cases.json")}
		if mode == "replay" {
			args = append(args, "--construction", filepath.Join(work, "caller-fill-rejection-construct", "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--construction-cases", filepath.Join(jointFillRejectionRoot, "construction-cases.json"),
				"--attempts", strconv.Itoa(budget), "--out", directory)
		}
		raw, runErr := commandOutput(input.Root, nil, binary, args...)
		if len(raw) > 0 {
			name := input.Target.ID + "-caller-fill-rejection-" + mode + ".json"
			if err := os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0o644); err != nil {
				return err
			}
		}
		if runErr != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_FILL_REJECTION: %w", runErr)
		}
		selected, err = validateJointFillRejectionSmoke(raw, source, feedback, evaluation, budget, mode == "replay", selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CALLER_FILL_REJECTION: %w", err)
		}
	}
	return nil
}
