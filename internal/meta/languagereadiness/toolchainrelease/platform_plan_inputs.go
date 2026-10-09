package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type planInputSmokeCase struct {
	languageSmokeCase
	inputs, model string
}

func planInputExamples() []planInputSmokeCase {
	return []planInputSmokeCase{
		{languageSmokeCase{"record-korean", scalarIdentityRoot + "/source.gooo.fixture",
			scalarIdentityRoot + "/cases.json", "Describe", 4}, "examples/composition-inputs/record.json",
			scalarIdentityRoot + "/model/model.json"},
		{languageSmokeCase{"record-english", scalarIdentityRoot + "/english.gooo.fixture",
			scalarIdentityRoot + "/cases.json", "Describe", 4}, "examples/composition-inputs/record.json", ""},
		{languageSmokeCase{"input-joins", "examples/body-codegen/native-input-joins.gooo.fixture",
			"examples/body-codegen/native-input-joins-cases.json", "", 49}, "examples/composition-inputs/joins.json", ""},
		{languageSmokeCase{"called-helpers", "examples/dependent-continuation/main.gooo.fixture",
			"examples/dependent-continuation/cases.json", "Main", 2}, "examples/composition-inputs/helpers.json", ""},
	}
}

func smokePlanInputs(binary, work string, input BuildInput) error {
	for _, example := range planInputExamples() {
		if err := smokePlanInputExample(binary, work, input, example); err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PLAN_INPUTS %s: %w", example.name, err)
		}
	}
	return nil
}

func smokePlanInputExample(binary, work string, input BuildInput, example planInputSmokeCase) error {
	source, err := os.ReadFile(filepath.Join(input.Root, example.source))
	if err != nil {
		return err
	}
	run := func(mode string, args []string) ([]byte, error) {
		raw, err := commandOutput(input.Root, nil, binary, args...)
		return retainLanguageCommandOutput(input, example.name+"-"+mode, raw, err)
	}
	raw, err := run("plan", planInputArgs(example, "--json"))
	if err != nil {
		return err
	}
	plan, err := validatePlanInputInspection(raw, source, example)
	if err != nil {
		return err
	}
	raw, err = run("template", planInputArgs(example, "--inputs-template"))
	if err != nil {
		return err
	}
	if err := validatePlanInputTemplate(raw, plan); err != nil {
		return err
	}
	return smokePlanInputExecutions(run, work, input, example, plan)
}

func planInputArgs(example planInputSmokeCase, mode string) []string {
	args := []string{"body-plan", "--source", example.source, mode}
	if example.entry != "" {
		args = append(args, "--entry", example.entry)
	}
	return args
}

func smokePlanInputExecutions(run func(string, []string) ([]byte, error), work string,
	input BuildInput, example planInputSmokeCase, plan bodyexecution.CompositionInspection) error {
	directory, selected := filepath.Join(work, "plan-inputs-"+example.name), ""
	for _, mode := range []string{"inputs", "inputs-replay", "scored-replay"} {
		raw, err := run(mode, planInputExecutionArgs(example, directory, mode))
		if err != nil {
			return err
		}
		selected, err = validatePlanInputExecution(raw, plan, example, mode, selected)
		if err != nil {
			return err
		}
	}
	actual, err := os.ReadFile(filepath.Join(directory, "inputs.json"))
	original, originalErr := os.ReadFile(filepath.Join(input.Root, example.inputs))
	if err != nil || originalErr != nil || string(actual) != string(original) {
		return fmt.Errorf("input document was not retained byte for byte: %v %v", err, originalErr)
	}
	return nil
}

func planInputExecutionArgs(example planInputSmokeCase, directory, mode string) []string {
	source := example.source
	if mode != "inputs" {
		source = filepath.Join(directory, "original.gooo")
	}
	args := []string{"body-compose", "--source", source}
	if mode == "scored-replay" {
		args = append(args, "--cases", example.cases)
	} else {
		args = append(args, "--inputs", example.inputs)
	}
	if mode != "inputs" {
		return append(args, "--composition", filepath.Join(directory, "composition.json"))
	}
	if example.entry != "" {
		args = append(args, "--entry", example.entry)
	}
	if example.model != "" {
		args = append(args, "--model", example.model)
	}
	return append(args, "--repeat", "2", "--out", directory)
}
