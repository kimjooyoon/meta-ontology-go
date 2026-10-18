package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
)

var computationProfiles = []packageProfile{
	{"string-operations", "examples/bounded-string-operations", "app/stringops", "Main", "gooo://stringops/app/main", 20},
	{"single-condition-byte", "examples/single-condition-byte-search", "gooo/examples/bytefill", "IndexByte", "gooo://bytefill/index-byte", 10},
	{"collection-map", "examples/collection-map", "example/labels", "Classify", "urn:collection-map:classify", 2},
}

func smokeComputationPackages(binary string, input BuildInput) error {
	for _, profile := range computationProfiles {
		if err := smokeComputationPackage(binary, input, profile); err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_COMPUTATION %s: %w", profile.Name, err)
		}
	}
	return nil
}

func smokeComputationPackage(binary string, input BuildInput, profile packageProfile) error {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(input.Root, profile.Root, name))
	}
	manifest, err := read("gooo.workspace.json")
	if err != nil {
		return err
	}
	cases, err := read("cases.json")
	if err != nil {
		return err
	}
	program, err := computationPackageProgram(filepath.Join(input.Root, profile.Root), manifest)
	if err != nil {
		return err
	}
	receipt := filepath.Join(input.OutputDir, input.Target.ID+"-"+profile.Name+"-execute.json")
	selected := boundedTextSmokeBinding{}
	for _, replay := range []bool{false, true} {
		args := []string{"package", "execute", "--json", "--cases", filepath.Join(profile.Root, "cases.json")}
		output := receipt
		if replay {
			args[1] = "replay"
			args = append(args, "--receipt", receipt)
			output = filepath.Join(input.OutputDir, input.Target.ID+"-"+profile.Name+"-replay.json")
		}
		raw, runErr := commandOutput(input.Root, nil, binary, append(args, filepath.Join(profile.Root, "gooo.workspace.json"))...)
		if len(raw) > 0 {
			if err := os.WriteFile(output, raw, 0o644); err != nil {
				return err
			}
		}
		if runErr != nil {
			return runErr
		}
		selected, err = validateComputationPackage(raw, manifest, cases, program, profile, replay, selected)
		if err != nil {
			return err
		}
	}
	return nil
}
