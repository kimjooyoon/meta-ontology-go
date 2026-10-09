package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

const typedPathsRoot = "examples/caller-typed-paths"
const typedPathsModel = scalarIdentityRoot + "/model/model.json"

func smokeTypedPaths(binary, work string, input BuildInput) error {
	contextSHA, err := smokeTypedPreflight(binary, input)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(filepath.Join(input.Root, typedPathsRoot, "mixed-model.gooo.fixture"))
	if err != nil {
		return err
	}
	feedback, err := os.ReadFile(filepath.Join(input.Root, typedPathsRoot, "mixed-construction-cases.json"))
	if err != nil {
		return err
	}
	cases, err := os.ReadFile(filepath.Join(input.Root, typedPathsRoot, "mixed-evaluation-cases.json"))
	if err != nil {
		return err
	}
	directory, selected := filepath.Join(work, "typed-paths"), ""
	for _, replay := range []bool{false, true} {
		raw, err := runTypedSmoke(binary, directory, input, replay)
		if err != nil {
			return err
		}
		selected, err = validateTypedSmoke(raw, source, feedback, cases, contextSHA, replay, selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_TYPED_PATHS: %w", err)
		}
	}
	return nil
}

func runTypedSmoke(binary, directory string, input BuildInput, replay bool) ([]byte, error) {
	source, mode := filepath.Join(typedPathsRoot, "mixed-model.gooo.fixture"), "construct"
	if replay {
		source, mode = filepath.Join(directory, "original.gooo"), "replay"
	}
	args := []string{"body-construct", "--source", source,
		"--cases", filepath.Join(typedPathsRoot, "mixed-evaluation-cases.json")}
	if replay {
		args = append(args, "--construction", filepath.Join(directory, "construction.json"))
	} else {
		args = append(args, "--entry", "Main", "--model", typedPathsModel, "--attempts", "16",
			"--construction-cases", filepath.Join(typedPathsRoot, "mixed-construction-cases.json"), "--out", directory)
	}
	raw, err := commandOutput(input.Root, nil, binary, args...)
	return retainLanguageCommandOutput(input, "typed-path-"+mode, raw, err)
}
