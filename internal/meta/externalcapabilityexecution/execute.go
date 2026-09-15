package externalcapabilityexecution

import (
	"bytes"
	"os"
	"path/filepath"
)

func executeCapabilities(workspace, externalRoot string) ([]CapabilityRun, int, error) {
	tools, err := prepareTools(workspace, externalRoot)
	if err != nil {
		return nil, 0, err
	}
	runs := make([]CapabilityRun, 0, 2)
	for index := 1; index <= 2; index++ {
		run, err := executeRun(workspace, externalRoot, tools, index)
		if err != nil {
			return runs, len(runs) * 2, err
		}
		runs = append(runs, run)
	}
	return runs, 4, nil
}

func executeRun(workspace, externalRoot string, tools capabilityTools, index int) (CapabilityRun, error) {
	evalCode, evalOutput, err := runCommand("", nil, tools.Evaluator)
	if err != nil {
		return CapabilityRun{}, err
	}
	evaluated, evalJSON := decodeEvaluator(evalOutput)
	runRoot := filepath.Join(workspace, "macro", string(rune('0'+index)))
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		return CapabilityRun{}, err
	}
	input, err := os.ReadFile(filepath.Join(externalRoot, "_example", "make_fibonacci.gomacro"))
	if err != nil {
		return CapabilityRun{}, err
	}
	expected, err := os.ReadFile(filepath.Join(externalRoot, "_example", "make_fibonacci.gomacro_output"))
	if err != nil {
		return CapabilityRun{}, err
	}
	if err := os.WriteFile(filepath.Join(runRoot, "make_fibonacci.gomacro"), input, 0o644); err != nil {
		return CapabilityRun{}, err
	}
	macroCode, _, err := runCommand(runRoot, nil, tools.Gomacro, "-m", "-w", "make_fibonacci.gomacro")
	if err != nil {
		return CapabilityRun{}, err
	}
	generated, _ := os.ReadFile(filepath.Join(runRoot, "make_fibonacci.go"))
	run := CapabilityRun{
		RunID: "run-" + string(rune('0'+index)), Arithmetic: evaluated.Arithmetic,
		Function: evaluated.Function, EvaluatorExitCode: evalCode,
		EvaluatorOutputBytes: len(evalOutput), EvaluatorOutputSHA256: digestBytes(evalOutput),
		MacroExitCode:        macroCode,
		MacroGeneratedSHA256: digestBytes(generated), MacroExpectedSHA256: digestBytes(expected),
	}
	exact := evalCode == 0 && evalJSON && run.Arithmetic == "42" && run.Function == "55"
	exact = exact && macroCode == 0 && len(generated) > 0 && bytes.Equal(generated, expected)
	run.Status = status(exact)
	run.NormalizedSHA256 = normalizedRunDigest(run)
	return run, nil
}
