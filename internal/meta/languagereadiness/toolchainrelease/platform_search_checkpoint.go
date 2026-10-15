package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

const searchCheckpointRoot = "examples/caller-search-checkpoint"

func smokeSearchCheckpoint(binary, work string, input BuildInput) error {
	files := make([][]byte, 3)
	for i, name := range []string{"main.gooo.fixture", "construction-cases.json", "evaluation-cases.json"} {
		raw, err := os.ReadFile(filepath.Join(input.Root, searchCheckpointRoot, name))
		if err != nil {
			return err
		}
		files[i] = raw
	}
	directory := filepath.Join(work, "search-checkpoint")
	retained := ""
	for _, replay := range []bool{false, true} {
		args := []string{"body-construct", "--source", filepath.Join(searchCheckpointRoot, "main.gooo.fixture"),
			"--cases", filepath.Join(searchCheckpointRoot, "evaluation-cases.json"), "--format", "json"}
		mode := "construct"
		if replay {
			mode = "replay"
			args = append(args, "--construction", filepath.Join(directory, "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--attempts", "8", "--construction-cases",
				filepath.Join(searchCheckpointRoot, "construction-cases.json"), "--out", directory)
		}
		raw, err := commandOutput(input.Root, nil, binary, args...)
		if len(raw) > 0 {
			path := filepath.Join(input.OutputDir, input.Target.ID+"-search-checkpoint-"+mode+".json")
			if writeErr := os.WriteFile(path, raw, 0o644); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_SEARCH_CHECKPOINT: %w", err)
		}
		retained, err = validateSearchCheckpoint(raw, files[0], files[1], files[2], replay, retained)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_SEARCH_CHECKPOINT: %w", err)
		}
	}
	return nil
}
