package toolchainrelease

import (
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

const relativeV3Root = "examples/body-codegen/source-output-relations-v3-"

// New zero-weight release regression; original fitted study weights are not read.
func smokeRelativeV3(binary, work string, input BuildInput) error {
	m, err := contractdecision.NewSourceRelativeV3([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		return err
	}
	raw, err := m.Marshal()
	if err != nil {
		return err
	}
	model := filepath.Join(work, "relative-v3-zero.json")
	if err = os.WriteFile(model, raw, 0600); err != nil {
		return err
	}
	source, err := os.ReadFile(filepath.Join(input.Root, relativeV3Root+"cases.gooo.fixture"))
	if err != nil {
		return err
	}
	feedback, err := os.ReadFile(filepath.Join(input.Root, relativeV3Root+"construction.json"))
	if err != nil {
		return err
	}
	cases, err := os.ReadFile(filepath.Join(input.Root, relativeV3Root+"evaluation.json"))
	if err != nil {
		return err
	}
	observed, err := commandOutput(input.Root, nil, binary, "body-context", "--activity", "Choose", "--model", model, relativeV3Root+"cases.gooo.fixture")
	observed, err = retainLanguageCommandOutput(input, "relative-v3-preflight", observed, err)
	if err != nil {
		return err
	}
	preflight, err := validateRelativeV3Preflight(observed, source, raw)
	if err != nil {
		return err
	}
	directory, selected := filepath.Join(work, "relative-v3-construction"), ""
	for _, replay := range []bool{false, true} {
		name, mode := relativeV3Root+"cases.gooo.fixture", "construct"
		if replay {
			name, mode = filepath.Join(directory, "original.gooo"), "replay"
		}
		args := []string{"body-construct", "--source", name, "--cases", relativeV3Root + "evaluation.json"}
		if replay {
			args = append(args, "--construction", filepath.Join(directory, "construction.json"))
		} else {
			args = append(args, "--entry", "Main", "--model", model, "--attempts", "4", "--construction-cases", relativeV3Root+"construction.json", "--out", directory)
		}
		observed, err = commandOutput(input.Root, nil, binary, args...)
		observed, err = retainLanguageCommandOutput(input, "relative-v3-"+mode, observed, err)
		if err != nil {
			return err
		}
		selected, err = validateRelativeV3Smoke(observed, source, feedback, cases, preflight, input.ExpectedHead, replay, selected)
		if err != nil {
			return err
		}
		if !replay {
			if err = os.Remove(model); err != nil {
				return err
			}
		}
	}
	return nil
}
