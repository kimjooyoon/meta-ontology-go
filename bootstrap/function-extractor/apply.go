package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
)

func stageExtractions(root string, plans map[string]planSubject, residual []string,
	recipes []extractionRecipe) (map[string]stagedFile, []extractionSubject, []string, error) {
	bySubject := make(map[string]extractionRecipe, len(recipes))
	for _, recipe := range recipes {
		if _, exists := bySubject[recipe.Subject]; exists {
			return nil, nil, nil, fmt.Errorf("duplicate recipe %s", recipe.Subject)
		}
		bySubject[recipe.Subject] = recipe
	}
	buffers := make(map[string][]byte)
	changedBySubject := make(map[string][]string)
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
	}
	staged, err := formatStaged(root, buffers)
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
		source, exists := staged[logical]
		if !exists {
			return nil, nil, nil, fmt.Errorf("recipe did not rewrite subject %s", logical)
		}
		subjects = append(subjects, extractionSubject{
			Logical: logical, Before: plans[logical].Lines, After: extractionLines(source.data),
			Files: files, Consumer: "function-extractor",
			Operation: bySubject[logical].Operation, Proof: "axiomatic-foundation",
		})
	}
	return staged, subjects, unhandled, nil
}
