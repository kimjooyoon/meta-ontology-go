package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

const boundedTextRoot = "examples/bounded-text-search"

func smokeBoundedText(binary string, input BuildInput) error {
	manifest, err := os.ReadFile(filepath.Join(input.Root, boundedTextRoot, "gooo.workspace.json"))
	if err != nil {
		return err
	}
	cases, err := os.ReadFile(filepath.Join(input.Root, boundedTextRoot, "cases.json"))
	if err != nil {
		return err
	}
	receipt := filepath.Join(input.OutputDir, input.Target.ID+"-bounded-text-execute.json")
	selected := boundedTextSmokeBinding{}
	for _, replay := range []bool{false, true} {
		mode, output := "execute", receipt
		args := []string{"package", mode, "--json", "--cases", filepath.Join(boundedTextRoot, "cases.json")}
		if replay {
			args[1] = "replay"
			args = append(args, "--receipt", receipt)
			output = filepath.Join(input.OutputDir, input.Target.ID+"-bounded-text-replay.json")
		}
		raw, err := commandOutput(input.Root, nil, binary, append(args, filepath.Join(boundedTextRoot, "gooo.workspace.json"))...)
		if len(raw) > 0 {
			if writeErr := os.WriteFile(output, raw, 0o644); writeErr != nil {
				return writeErr
			}
		}
		if err == nil {
			selected, err = validateBoundedTextSmoke(raw, manifest, cases, replay, selected)
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_BOUNDED_TEXT: %w", err)
		}
	}
	return nil
}
