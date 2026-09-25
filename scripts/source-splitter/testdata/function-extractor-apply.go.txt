package main

import (
	"bytes"
	"fmt"
	"os"

	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

type stagedFile struct {
	name    string
	data    []byte
	mode    uint32
	created bool
}

func indexRecipes(recipes []extractionRecipe) (map[string]extractionRecipe, error) {
	return recipeauthority.Index(recipes)
}

func stageExtractions(root string, plans map[string]planSubject, residual []string,
	recipes []extractionRecipe) (map[string]stagedFile, []extractionSubject, []string, []extractionFailureRecord, error) {
	bySubject, err := indexRecipes(recipes)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	buffers := make(map[string][]byte)
	created := make(map[string]bool)
	changedBySubject := make(map[string][]string)
	createdBySubject := make(map[string][]string)
	operationsBySubject := make(map[string][]string)
	evidenceBySubject := make(map[string][]projectionextractor.StrategyEvidence)
	unhandled := make([]string, 0)
	failures := make([]extractionFailureRecord, 0)
	for _, logical := range residual {
		recipe, exists := bySubject[logical]
		if !exists {
			operations, evidence, err := stageGenericExtraction(root, logical, buffers, created, changedBySubject, createdBySubject)
			if err != nil {
				unhandled = append(unhandled, logical)
				failures = append(failures, extractionFailure(logical, err))
				fmt.Printf("function-extractor: unhandled=%s %v\n", logical, err)
				continue
			}
			operationsBySubject[logical] = operations
			evidenceBySubject[logical] = evidence
			continue
		}
		for _, edit := range recipe.Edits {
			name, err := extractionPath(root, edit.Path)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			data, exists := buffers[edit.Path]
			if !exists {
				data, err = os.ReadFile(name)
				if err != nil {
					return nil, nil, nil, nil, err
				}
			}
			oldText, newText := editText(edit.Old), editText(edit.New)
			if bytes.Count(data, oldText) != 1 {
				return nil, nil, nil, nil, fmt.Errorf("recipe %s is not exact in %s", logical, edit.Path)
			}
			buffers[edit.Path] = bytes.Replace(data, oldText, newText, 1)
			changedBySubject[logical] = appendUnique(changedBySubject[logical], edit.Path)
		}
		paths, err := stageCreations(root, recipe, buffers, created)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		for _, path := range paths {
			changedBySubject[logical] = appendUnique(changedBySubject[logical], path)
			createdBySubject[logical] = appendUnique(createdBySubject[logical], path)
		}
	}
	staged, err := formatStaged(root, buffers, created)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	subjects, err := extractionSubjects(plans, residual, bySubject, operationsBySubject, evidenceBySubject, changedBySubject, createdBySubject, staged)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return staged, subjects, unhandled, failures, nil
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
