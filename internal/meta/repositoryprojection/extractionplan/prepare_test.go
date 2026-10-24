package extractionplan

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func TestPreparationKeepsEarlierFailureWhenLaterRecipeCannotStage(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"go.mod": "module example.test\n",
		"a.go": "package fixture\nfunc Broken(", "b.go": "package fixture\nfunc Value() int { return 1 }\n"}
	for path, source := range files {
		if err := os.WriteFile(filepath.Join(root, path), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	recipe := recipeauthority.ExtractionRecipe{Subject: "b.go", Operation: "extract-function",
		Edits: []recipeauthority.TextEdit{{Path: "b.go", Old: []string{"return 7"}, New: []string{"return 2"}}}}
	result, err := Prepare(root, []string{"a.go", "b.go"}, []recipeauthority.ExtractionRecipe{recipe})
	if err == nil || len(result.Files) != 0 {
		t.Fatal("failed preparation became staged output", result, err)
	}
	if earlier, ok := errors.AsType[projectionextractor.Failure](result.Failures["a.go"]); !ok || earlier.Reason != "PARSER_EVIDENCE_MISSING" {
		t.Fatal("later recipe failure erased the earlier parser observation", result.Failures)
	}
	for path, source := range files {
		if raw, err := os.ReadFile(filepath.Join(root, path)); err != nil || string(raw) != source {
			t.Fatal("preparation modified source", path, err)
		}
	}
}
