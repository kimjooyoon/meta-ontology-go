package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
)

func stageExtractions(root string, plans map[string]planSubject, residual []string,
	recipes []extractionRecipe) (map[string]stagedFile, []extractionSubject, []string, error) {
	bySubject, err := indexRecipes(recipes)
	if err != nil {
		return nil, nil, nil, err
	}
	buffers := make(map[string][]byte)
	created := make(map[string]bool)
	changedBySubject := make(map[string][]string)
	createdBySubject := make(map[string][]string)
	unhandled := make([]string, 0)
	for _, logical := range residual {
		recipe, exists := bySubject[logical]
		if !exists {
			unhandled = append(unhandled, logical)
			continue
		}
		for _, edit := range recipe.Edits {
			name, err := extractionPath(root, edit.Path)
			if err != nil {
				return nil, nil, nil, err
			}
			data, exists := buffers[edit.Path]
			if !exists {
				data, err = os.ReadFile(name)
				if err != nil {
					return nil, nil, nil, err
				}
			}
			oldText, newText := editText(edit.Old), editText(edit.New)
			if bytes.Count(data, oldText) != 1 {
				return nil, nil, nil, fmt.Errorf("recipe %s is not exact in %s", logical, edit.Path)
			}
			buffers[edit.Path] = bytes.Replace(data, oldText, newText, 1)
			changedBySubject[logical] = appendUnique(changedBySubject[logical], edit.Path)
		}
		paths, err := stageCreations(root, recipe, buffers, created)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, path := range paths {
			changedBySubject[logical] = appendUnique(changedBySubject[logical], path)
			createdBySubject[logical] = appendUnique(createdBySubject[logical], path)
		}
	}
	staged, err := formatStaged(root, buffers, created)
	if err != nil {
		return nil, nil, nil, err
	}
	subjects := make([]extractionSubject, 0, len(residual)-len(unhandled))
	for _, logical := range residual {
		files := changedBySubject[logical]
		if len(files) == 0 {
			continue
		}
		sort.Strings(files)
		createdFiles := createdBySubject[logical]
		sort.Strings(createdFiles)
		source, exists := staged[logical]
		if !exists {
			return nil, nil, nil, fmt.Errorf("recipe did not rewrite subject %s", logical)
		}
		subjects = append(subjects, extractionSubject{
			Logical: logical, Before: plans[logical].Lines, After: extractionLines(source.data),
			Files: files, CreatedFiles: createdFiles, Consumer: "function-extractor",
			Operation: bySubject[logical].Operation, Proof: "axiomatic-foundation",
		})
	}
	return staged, subjects, unhandled, nil
}
