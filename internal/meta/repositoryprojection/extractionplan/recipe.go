package extractionplan

import (
	"bytes"
	"fmt"
	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	"os"
)

type extractionRecipe = recipeauthority.ExtractionRecipe

func stageRecipe(root, logical string, recipe extractionRecipe, buffers map[string][]byte,
	created map[string]bool, changedBySubject, createdBySubject map[string][]string) error {
	if err := validateRecipeModule(root, logical, recipe); err != nil {
		return err
	}
	for _, edit := range recipe.Edits {
		name, err := extractionPath(root, edit.Path)
		if err != nil {
			return err
		}
		data, exists := buffers[edit.Path]
		if !exists {
			data, err = os.ReadFile(name)
			if err != nil {
				return err
			}
		}
		oldText, newText := editText(edit.Old), editText(edit.New)
		if bytes.Count(data, oldText) != 1 {
			return fmt.Errorf("recipe %s is not exact in %s", logical, edit.Path)
		}
		buffers[edit.Path] = bytes.Replace(data, oldText, newText, 1)
		changedBySubject[logical] = appendUnique(changedBySubject[logical], edit.Path)
	}
	paths, err := stageCreations(root, recipe, buffers, created)
	if err != nil {
		return err
	}
	for _, path := range paths {
		changedBySubject[logical] = appendUnique(changedBySubject[logical], path)
		createdBySubject[logical] = appendUnique(createdBySubject[logical], path)
	}
	for _, path := range changedBySubject[logical] {
		if err := projectionextractor.ValidateSourceImports(root, path, buffers[path]); err != nil {
			return err
		}
	}
	return nil
}

func stageCreations(root string, recipe extractionRecipe, buffers map[string][]byte,
	created map[string]bool) ([]string, error) {
	paths := make([]string, 0, len(recipe.Creates))
	for _, creation := range recipe.Creates {
		name, err := extractionPath(root, creation.Path)
		if err != nil {
			return nil, err
		}
		if len(creation.Lines) == 0 || buffers[creation.Path] != nil {
			return nil, fmt.Errorf("invalid creation %s", creation.Path)
		}
		if _, err := os.Lstat(name); err == nil || !os.IsNotExist(err) {
			return nil, fmt.Errorf("creation target exists: %s", creation.Path)
		}
		buffers[creation.Path] = editText(creation.Lines)
		created[creation.Path] = true
		paths = append(paths, creation.Path)
	}
	return paths, nil
}
