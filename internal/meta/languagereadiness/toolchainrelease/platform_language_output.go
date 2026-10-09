package toolchainrelease

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func runLanguageSmokeCommand(binary, directory string, input BuildInput, example languageSmokeCase,
	replay bool, run func(string, []string, string, ...string) ([]byte, error)) ([]byte, error) {
	mode := "construct"
	if replay {
		mode = "replay"
	}
	raw, runErr := run(input.Root, nil, binary, languageSmokeArgs(example, directory, replay)...)
	extension := ".json"
	if runErr != nil {
		// Combined command output can include diagnostics or incomplete JSON.
		// Preserve its exact bytes without presenting it as a successful receipt.
		extension = ".failed-output"
	}
	name := input.Target.ID + "-" + example.name + "-" + mode + extension
	if err := os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0o644); err != nil {
		return raw, errors.Join(runErr, fmt.Errorf("retain language command output: %w", err))
	}
	return raw, runErr
}
