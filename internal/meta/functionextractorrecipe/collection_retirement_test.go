package recipe

import "testing"

func TestCollectionRegistryRecipeRetirement8559(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	const subject = "internal/meta/languagereadiness/languagesyntax/evaluate.go"
	active, err := Index(manifest.Recipes)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := active[subject]; found {
		t.Fatal("already moved helper still has executable move authority")
	}
	count := 0
	for _, retired := range manifest.RetiredRecipes {
		if retired.Recipe.Subject == subject {
			count++
			if retired.Recipe.Operation != "extract-language-syntax-registry-paths" ||
				retired.RetiredAt != "0d8a4bfb3afe33c028b9686cc410950a9a832f1f" ||
				retired.ReasonCode != "SOURCE_SPLIT_INVALIDATED_EXACT_PRECONDITION" || len(retired.Recipe.Edits) != 3 {
				t.Fatal("retired operation lost its original source identity")
			}
		}
	}
	if count != 1 {
		t.Fatal("retirement origin must be unique", count)
	}
	t.Log("8559: stale helper move retired with source and recipe digests; generic extraction remains available")
}
