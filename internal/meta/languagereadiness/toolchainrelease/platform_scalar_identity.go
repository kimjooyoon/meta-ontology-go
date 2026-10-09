package toolchainrelease

import (
	"fmt"
	"path/filepath"
)

const scalarIdentityRoot = "examples/scalar-identity"

func scalarIdentityExamples() []languageSmokeCase {
	return []languageSmokeCase{
		{"scalar-identity-korean", scalarIdentityRoot + "/source.gooo.fixture", scalarIdentityRoot + "/cases.json", "Describe", 4},
		{"scalar-identity-english", scalarIdentityRoot + "/english.gooo.fixture", scalarIdentityRoot + "/cases.json", "Describe", 4},
	}
}

func smokeScalarIdentity(binary, work string, input BuildInput) error {
	common := ""
	for i, example := range scalarIdentityExamples() {
		names := [3]string{"정수", "논리", "문자열"}
		if i == 1 {
			names = [3]string{"Count", "Enabled", "Message"}
		}
		for _, model := range []bool{false, true} {
			selected, err := smokeScalarIdentityMode(binary, work, input, example, names, model)
			if err != nil {
				return err
			}
			if common != "" && common != selected {
				return fmt.Errorf("scalar rename or model changed the generated program")
			}
			common = selected
		}
	}
	return nil
}

func smokeScalarIdentityMode(binary, work string, input BuildInput, example languageSmokeCase,
	names [3]string, model bool) (string, error) {
	if model {
		example.name += "-model"
	} else {
		example.name += "-fixed"
	}
	directory, selected := filepath.Join(work, example.name), ""
	contextSHA := ""
	if model {
		var err error
		contextSHA, err = smokeScalarModelPreflight(binary, input, example)
		if err != nil {
			return "", err
		}
	}
	for _, replay := range []bool{false, true} {
		run := func(root string, env []string, command string, args ...string) ([]byte, error) {
			if model && !replay {
				args = append(args, "--model", filepath.Join(scalarIdentityRoot, "model", "model.json"))
			}
			return commandOutput(root, env, command, args...)
		}
		raw, err := runLanguageSmokeCommand(binary, directory, input, example, replay, run)
		if err != nil {
			return "", fmt.Errorf("TOOLCHAIN_RELEASE_SCALAR_IDENTITY %s: %w", example.name, err)
		}
		selected, err = validateScalarIdentitySmoke(raw, names, model, replay, selected)
		if err != nil {
			return "", err
		}
		if model {
			if err := validateScalarPreflightBinding(raw, contextSHA); err != nil {
				return "", err
			}
		}
	}
	return selected, nil
}
