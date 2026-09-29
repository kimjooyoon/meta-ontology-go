package recipe

import (
	"strings"
	"testing"
)

func TestLoadManifestRetainsRetiredRecipeAsNonExecutableEvidence(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != manifestSchema || len(manifest.RetiredRecipes) != 1 {
		t.Fatalf("manifest identity = %q with %d retired recipes", manifest.Schema, len(manifest.RetiredRecipes))
	}
	retired := manifest.RetiredRecipes[0]
	if retired.Recipe.Subject != "cmd/gooo/main_part01.go" ||
		retired.Recipe.Operation != "extract-cli-analysis-types" ||
		retired.RecipeDigest != "sha256:ee577a3384488b2cbbd62fd6a57b1eef756d19ff0137cf1fc8b4a5e53e77dfe6" ||
		retired.SourceDigest != "sha256:a1effbc73c1717412fccb6f172d5d689e9c9dc26f9b803dd55b59b677bb9f401" ||
		retired.RetiredAt != "c21c26c61172bf8a56cc7929ac31dc3ad419c021" ||
		retired.ReasonCode != "SOURCE_SPLIT_INVALIDATED_EXACT_PRECONDITION" {
		t.Fatalf("retirement evidence = %#v", retired)
	}
	active, err := Index(manifest.Recipes)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := active[retired.Recipe.Subject]; ok {
		t.Fatalf("retired subject %q remains executable", retired.Recipe.Subject)
	}
}

func TestValidateManifestRejectsRetiredRecipePayloadMutation(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.RetiredRecipes[0].Recipe.Edits[0].Old[0] += " "
	err = validateManifest(manifest)
	if err == nil || !strings.Contains(err.Error(), "digest does not match archived payload") {
		t.Fatalf("validate mutated retirement payload error = %v", err)
	}
}

func TestValidateManifestRejectsRecipeThatIsBothActiveAndRetired(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Recipes = append(manifest.Recipes, manifest.RetiredRecipes[0].Recipe)
	err = validateManifest(manifest)
	if err == nil || !strings.Contains(err.Error(), "both active and retired") {
		t.Fatalf("validate active/retired overlap error = %v", err)
	}
}
