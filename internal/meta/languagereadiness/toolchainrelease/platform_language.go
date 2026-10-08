package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type languageSmokeCase struct {
	name, source, cases, entry string
	total                      int
}

func smokeLanguageExamples(binary, work string, input BuildInput) error {
	examples := []languageSmokeCase{
		{"integer-division", "examples/integer-division/source.gooo.fixture", "examples/integer-division/cases.json", "", 8},
		{"candidate-locals", "examples/candidate-locals/retry.gooo.fixture", "examples/candidate-locals/cases.json", "", 12},
		{"text-operations", "examples/text-operations/source.gooo.fixture", "examples/text-operations/cases.json", "Classify", 12},
	}
	if err := os.MkdirAll(input.OutputDir, 0o755); err != nil {
		return err
	}
	for _, example := range examples {
		if err := smokeLanguageExample(binary, work, input, example); err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_LANGUAGE_SMOKE %s: %w", example.name, err)
		}
	}
	if err := smokeSourceGraph(binary, input); err != nil {
		return err
	}
	if err := smokePackageText(binary, input); err != nil {
		return err
	}
	if err := smokePackageSource(binary, input); err != nil {
		return err
	}
	if err := smokeJointConstruction(binary, work, input); err != nil {
		return err
	}
	return smokeJointSearchRejection(binary, work, input)
}

func smokeLanguageExample(binary, work string, input BuildInput, example languageSmokeCase) error {
	directory := filepath.Join(work, "language-"+example.name)
	selected := ""
	for _, replay := range []bool{false, true} {
		mode := "construct"
		if replay {
			mode = "replay"
		}
		args := languageSmokeArgs(example, directory, replay)
		raw, err := commandOutput(input.Root, nil, binary, args...)
		if err != nil {
			return err
		}
		name := input.Target.ID + "-" + example.name + "-" + mode + ".json"
		if err := os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0o644); err != nil {
			return err
		}
		observed, err := validateLanguageSmoke(raw, example.total, replay, selected)
		if err != nil {
			return err
		}
		selected = observed
	}
	return nil
}

func languageSmokeArgs(example languageSmokeCase, directory string, replay bool) []string {
	source := example.source
	if replay {
		source = filepath.Join(directory, "original.gooo")
	}
	args := []string{"body-compose", "--source", source, "--cases", example.cases}
	if replay {
		return append(args, "--composition", filepath.Join(directory, "composition.json"))
	}
	if example.entry != "" {
		args = append(args, "--entry", example.entry)
	}
	return append(args, "--out", directory)
}

func validateLanguageSmoke(raw []byte, expected int, replay bool, selected string) (string, error) {
	var result struct {
		Generated *bool `json:"generated_now"`
		Runtime   struct {
			Passed     int  `json:"finite_passed"`
			Total      int  `json:"finite_total"`
			ModelCalls *int `json:"model_calls"`
			Projection bool `json:"projection_replayed"`
			Replay     bool `json:"runtime_replayed"`
		} `json:"runtime"`
		Composition struct {
			GeneratedSHA256 string `json:"generated_sha256"`
		} `json:"composition"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	r := result.Runtime
	if expected < 1 || result.Generated == nil || *result.Generated == replay ||
		r.Passed != expected || r.Total != expected || r.ModelCalls == nil || *r.ModelCalls != 0 ||
		!r.Projection || !r.Replay || result.Composition.GeneratedSHA256 == "" {
		return "", fmt.Errorf("native language cases or saved replay differ")
	}
	if replay && (selected == "" || selected != result.Composition.GeneratedSHA256) {
		return "", fmt.Errorf("saved replay changed the generated program")
	}
	return result.Composition.GeneratedSHA256, nil
}
