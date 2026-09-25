package metriccounterfactual

import (
	"fmt"
	"os"
	"path/filepath"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func ProjectRootPolicy() RootPolicy {
	return RootPolicy{
		CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE",
		TopologyReason: "ROOT_TOPOLOGY_EXEMPT", ReadmeRequirement: "NOT_APPLICABLE",
	}
}

func BaselineManifest() (Manifest, error) {
	return SealManifest(Manifest{
		Schema: ManifestSchema,
		Files: []FileSpec{
			{Path: "logic/rules.gooo", Language: "gooo", Content: "entity source.\nrule derive(source).\n"},
			{Path: "runtime/main.go", Language: "go", Content: "package runtime\n\nfunc MainValue() int { return 1 }\n"},
			{Path: "runtime/nested/existing.go", Language: "go", Content: "package nested\n\nvar Existing = 1\n"},
		},
	})
}

func CounterfactualPlan() (Plan, error) {
	return SealPlan(Plan{
		Schema: PlanSchema,
		Mutations: []Mutation{
			{Kind: "APPEND", Path: "logic/rules.gooo", Content: "derive improved.\n"},
			{Kind: "CREATE", Path: "generated/deeper/new.go", Content: "package generated\n\nfunc Value() int { return 2 }\n"},
		},
	})
}

func Materialize(root string, manifest Manifest) error {
	if manifest.Schema != ManifestSchema || !ValidManifest(manifest) {
		return fmt.Errorf("invalid manifest")
	}
	seen := make(map[string]bool, len(manifest.Files))
	for _, file := range manifest.Files {
		if seen[file.Path] || languageForPath(file.Path) != file.Language {
			return fmt.Errorf("invalid manifest file %q", file.Path)
		}
		seen[file.Path] = true
		native, err := artifact.SafeNative(root, file.Path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(native), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(native, []byte(file.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
