package main

import (
	"bytes"
	"fmt"
	"os"
)

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
	unhandled := make([]string, 0)
	failures := make([]extractionFailureRecord, 0)
	for _, logical := range residual {
		recipe, exists := bySubject[logical]
		if !exists {
			if err := stageGenericExtraction(root, logical, buffers, created, changedBySubject, createdBySubject); err != nil {
				unhandled = append(unhandled, logical)
				failures = append(failures, extractionFailure(logical, err))
				fmt.Printf("function-extractor: unhandled=%s %v\n", logical, err)
				continue
			}
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
	subjects, err := extractionSubjects(plans, residual, bySubject, changedBySubject, createdBySubject, staged)
	if err != nil { return nil, nil, nil, nil, err }
	return staged, subjects, unhandled, failures, nil
}
