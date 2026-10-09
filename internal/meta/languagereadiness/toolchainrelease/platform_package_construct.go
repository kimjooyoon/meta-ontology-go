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
	Manifest, Feedback, Evaluation, Program []byte
	Source                                  string
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

func smokePackageConstruction(binary string, input BuildInput) error {
	reference, err := readPackageConstructionReference(input.Root)
	if err != nil {
		return err
	}
	var saved []byte
	for _, mode := range []string{"partial", "construct", "replay"} {
		budget := 6
		if mode == "partial" {
			budget = 5
		}
		args := []string{"package", "construct", "--json", "--cases", filepath.Join(packageConstructionRoot, "evaluation-cases.json")}
		if mode == "replay" {
			args = append(args, "--receipt", filepath.Join(input.OutputDir, input.Target.ID+"-package-caller-construct.json"))
		} else {
			args = append(args, "--construction-cases", filepath.Join(packageConstructionRoot, "construction-cases.json"), "--attempts", strconv.Itoa(budget))
		}
		raw, runErr := commandOutput(input.Root, nil, binary, append(args, filepath.Join(packageConstructionRoot, "gooo.workspace.json"))...)
		if len(raw) > 0 {
			if err := os.WriteFile(filepath.Join(input.OutputDir, input.Target.ID+"-package-caller-"+mode+".json"), raw, 0644); err != nil {
				return err
			}
		}
		if runErr != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_CONSTRUCTION %s: %w", mode, runErr)
		}
		if err := validatePackageConstructionSmoke(raw, reference, budget, saved); err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_CONSTRUCTION %s: %w", mode, err)
		}
		if mode == "construct" {
			saved = raw
		}
	}
	return nil
}
