package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

// Preparation reads source and lowers its graph without inference or execution.
func computationPackageProgram(root string, raw []byte) (workspaceexecution.Program, error) {
	var workspace struct {
		Schema   string
		Entry    packageruntime.EntrySpec
		Packages []struct {
			Path, Name       string
			Imports, Sources []string
		}
	}
	if err := json.Unmarshal(raw, &workspace); err != nil {
		return workspaceexecution.Program{}, err
	}
	if workspace.Schema != "gooo/package-workspace-manifest/v1" {
		return workspaceexecution.Program{}, fmt.Errorf("computation package manifest schema differs")
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: workspace.Entry}
	for _, p := range workspace.Packages {
		pkg := packageruntime.PackageSpec{Path: p.Path, Name: p.Name, Imports: p.Imports}
		for _, name := range p.Sources {
			raw, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				return workspaceexecution.Program{}, err
			}
			pkg.Sources = append(pkg.Sources, packageruntime.Source{Filename: name, Content: string(raw)})
		}
		manifest.Packages = append(manifest.Packages, pkg)
	}
	return workspaceexecution.Prepare(manifest)
}

// Compare the raw recorded program so absent, empty and null origins remain distinct.
func computationPackageReceiptProgramEqual(raw []byte, expected workspaceexecution.Program) bool {
	var recorded struct {
		Result struct{ Program json.RawMessage }
	}
	if json.Unmarshal(raw, &recorded) != nil {
		return false
	}
	current, err := json.Marshal(expected)
	return err == nil && samePackageProgramSource(recorded.Result.Program, current)
}

func computationDigest(raw []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
}
