package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const nativeArithmeticRoot = "examples/caller-native-failure"

func smokeNativeArithmetic(binary, work string, input BuildInput) error {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(input.Root, nativeArithmeticRoot, name))
	}
	source, err := read("source.gooo.fixture")
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
		budget := 6
		if mode == "partial" {
			budget = 5
		}
		args := []string{"body-construct", "--source", filepath.Join(nativeArithmeticRoot, "source.gooo.fixture"),
			"--cases", filepath.Join(nativeArithmeticRoot, "evaluation-cases.json")}
		if mode == "replay" {
			args = append(args, "--construction", filepath.Join(work, "native-arithmetic-construct", "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--construction-cases", filepath.Join(nativeArithmeticRoot, "construction-cases.json"),
				"--attempts", strconv.Itoa(budget), "--out", filepath.Join(work, "native-arithmetic-"+mode))
		}
		raw, err := runNativeArithmeticSmoke(binary, input, mode, args)
		if err != nil {
			return err
		}
		selected, err = validateNativeArithmeticSmoke(raw, source, feedback, evaluation, budget, mode == "replay", selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_NATIVE_ARITHMETIC: %w", err)
		}
	}
	return smokeNativeArithmeticGraph(binary, work, input)
}

func runNativeArithmeticSmoke(binary string, input BuildInput, mode string, args []string) ([]byte, error) {
	raw, runErr := commandOutput(input.Root, nil, binary, args...)
	if len(raw) > 0 {
		if err := os.WriteFile(filepath.Join(input.OutputDir, input.Target.ID+"-native-arithmetic-"+mode+".json"), raw, 0o644); err != nil {
			return nil, err
		}
	}
	if runErr != nil {
		return raw, fmt.Errorf("TOOLCHAIN_RELEASE_NATIVE_ARITHMETIC: %w", runErr)
	}
	return raw, nil
}
