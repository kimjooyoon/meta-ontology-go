package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

const helperFillRoot = "examples/caller-fill-history"

func smokeHelperFillHistory(binary, work string, input BuildInput) error {
	return runHelperFillHistory(binary, work, input, commandOutput)
}

func runHelperFillHistory(binary, work string, input BuildInput,
	run func(string, []string, string, ...string) ([]byte, error)) error {
	files := make([][]byte, 3)
	for i, name := range []string{"main.gooo.fixture", "construction-cases.json", "evaluation-cases.json"} {
		raw, err := os.ReadFile(filepath.Join(input.Root, helperFillRoot, name))
		if err != nil {
			return err
		}
		files[i] = raw
	}
	directory, retained := filepath.Join(work, "helper-fill-history"), ""
	for _, replay := range []bool{false, true} {
		args := []string{"body-construct", "--source", filepath.Join(helperFillRoot, "main.gooo.fixture"),
			"--cases", filepath.Join(helperFillRoot, "evaluation-cases.json"), "--format", "json"}
		mode := "construct"
		if replay {
			mode = "replay"
			args = append(args, "--construction", filepath.Join(directory, "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--attempts", "4", "--construction-cases",
				filepath.Join(helperFillRoot, "construction-cases.json"), "--out", directory)
		}
		raw, err := run(input.Root, nil, binary, args...)
		if len(raw) > 0 {
			path := filepath.Join(input.OutputDir, input.Target.ID+"-helper-fill-history-"+mode+".json")
			if writeErr := os.WriteFile(path, raw, 0o644); writeErr != nil {
				return writeErr
			}
		}
		if err == nil {
			retained, err = validateHelperFillHistory(raw, files[0], files[1], files[2], replay, retained)
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_HELPER_FILL_HISTORY: %w", err)
		}
	}
	return nil
}
