package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageConstructionRoot = "examples/package-caller-construction"

type packageConstructionReference struct {
	Manifest, Feedback, Evaluation, Inputs, Program []byte
	Source                                          string
}

// The reference is rebuilt from the original fixture, not from the executable's
// observation. It binds package identities, imported source bytes and lowering.
func readPackageConstructionReference(root string) (packageConstructionReference, error) {
	r := packageConstructionReference{}
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(root, packageConstructionRoot, name))
	}
	var err error
	if r.Manifest, err = read("gooo.workspace.json"); err != nil {
		return r, err
	}
	if r.Feedback, err = read("construction-cases.json"); err != nil {
		return r, err
	}
	if r.Evaluation, err = read("evaluation-cases.json"); err != nil {
		return r, err
	}
	if r.Inputs, err = read("inputs.json"); err != nil {
		return r, err
	}
	var m struct {
		Schema   string
		Entry    packageruntime.EntrySpec
		Packages []struct {
			Path, Name       string
			Imports, Sources []string
		}
	}
	if err = json.Unmarshal(r.Manifest, &m); err != nil {
		return r, err
	}
	if m.Schema != "gooo/package-workspace-manifest/v1" {
		return r, fmt.Errorf("package construction manifest schema differs")
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: m.Entry}
	for _, spec := range m.Packages {
		p := packageruntime.PackageSpec{Path: spec.Path, Name: spec.Name, Imports: spec.Imports}
		for _, name := range spec.Sources {
			b, err := read(name)
			if err != nil {
				return r, err
			}
			p.Sources = append(p.Sources, packageruntime.Source{Filename: name, Content: string(b)})
		}
		manifest.Packages = append(manifest.Packages, p)
	}
	program, err := workspaceexecution.Prepare(manifest)
	if err != nil {
		return r, err
	}
	r.Source = program.Source
	r.Program, err = json.Marshal(program)
	return r, err
}

type packageConstructionSmokeRun struct {
	mode      string
	budget    int
	inputOnly bool
	from      string
}

func smokePackageConstruction(binary string, input BuildInput) error {
	reference, err := readPackageConstructionReference(input.Root)
	if err != nil {
		return err
	}
	saved := map[string][]byte{}
	for _, step := range []packageConstructionSmokeRun{
		{"partial", 5, false, ""}, {"construct", 6, false, ""}, {"replay", 6, false, "construct"},
		{"partial-inputs", 5, true, ""}, {"inputs", 6, true, ""},
		{"inputs-replay", 6, true, "construct"}, {"inputs-again", 6, true, "inputs"},
		{"inputs-evaluate", 6, false, "inputs-again"},
	} {
		raw, err := runPackageConstructionSmoke(binary, input, step)
		if err == nil {
			err = validatePackageConstructionRun(raw, reference, step.budget, saved[step.from], step.inputOnly)
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_CONSTRUCTION %s: %w", step.mode, err)
		}
		saved[step.mode] = raw
	}
	return smokePackageInputValues(binary, input)
}

func packageConstructionSmokePath(input BuildInput, mode string) string {
	return filepath.Join(input.OutputDir, input.Target.ID+"-package-caller-"+mode+".json")
}

func runPackageConstructionSmoke(binary string, input BuildInput, step packageConstructionSmokeRun) ([]byte, error) {
	flag, name := "--cases", "evaluation-cases.json"
	if step.inputOnly {
		flag, name = "--inputs", "inputs.json"
	}
	args := []string{"package", "construct", "--json", flag, filepath.Join(packageConstructionRoot, name)}
	if step.from != "" {
		args = append(args, "--receipt", packageConstructionSmokePath(input, step.from))
	} else {
		args = append(args, "--construction-cases", filepath.Join(packageConstructionRoot, "construction-cases.json"), "--attempts", strconv.Itoa(step.budget))
	}
	raw, err := commandOutput(input.Root, nil, binary, append(args, filepath.Join(packageConstructionRoot, "gooo.workspace.json"))...)
	if len(raw) > 0 {
		if writeErr := os.WriteFile(packageConstructionSmokePath(input, step.mode), raw, 0o644); writeErr != nil {
			return raw, writeErr
		}
	}
	return raw, err
}

func smokePackageInputValues(binary string, input BuildInput) error {
	raw, err := commandOutput(input.Root, nil, binary, "package", "construct", "--receipt", packageConstructionSmokePath(input, "inputs-again"),
		"--inputs", filepath.Join(packageConstructionRoot, "inputs.json"), filepath.Join(packageConstructionRoot, "gooo.workspace.json"))
	if len(raw) > 0 {
		if writeErr := os.WriteFile(filepath.Join(input.OutputDir, input.Target.ID+"-package-caller-values.jsonl"), raw, 0o644); writeErr != nil {
			return writeErr
		}
	}
	if err != nil {
		return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_VALUES: %w", err)
	}
	if string(raw) != "3\n-10\n-9007199254740994\n1\n" {
		return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_VALUES: actual entry values differ")
	}
	return nil
}
