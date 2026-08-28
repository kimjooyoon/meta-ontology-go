package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
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
	operationBySubject := make(map[string]string)
	proofBySubject := make(map[string]string)
	unhandled := make([]string, 0)
	failures := make([]extractionFailureRecord, 0)
	for _, logical := range residual {
		recipe, exists := bySubject[logical]
		if !exists {
			generated, paths, extractionErr := genericASTExtraction(root, logical)
			if extractionErr != nil {
				unhandled = append(unhandled, logical)
				var failure extractionFailure
				if errors.As(extractionErr, &failure) {
					decision := "UNKNOWN"
					if failure.UnknownClass == "KNOWN_CONTRADICTION" {
						decision = "REFUTED"
					}
					failures = append(failures, extractionFailureRecord{Logical: logical, Decision: decision, Stage: failure.Stage, Step: failure.Step, Reason: failure.Reason, UnknownClass: failure.UnknownClass, NextOperation: failure.NextOperation, BlockedBy: failure.BlockedBy})
				}
				continue
			}
			for path, data := range generated {
				buffers[path] = data
				changedBySubject[logical] = appendUnique(changedBySubject[logical], path)
				if path != logical {
					created[path] = true
					createdBySubject[logical] = appendUnique(createdBySubject[logical], path)
				}
			}
			operationBySubject[logical] = "move-complete-declarations"
			proofBySubject[logical] = "coherent-system"
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
		operationBySubject[logical] = recipe.Operation
		proofBySubject[logical] = "axiomatic-foundation"
	}
	staged, err := formatStaged(root, buffers, created)
	if err != nil {
		return nil, nil, nil, nil, err
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
			return nil, nil, nil, nil, fmt.Errorf("recipe did not rewrite subject %s", logical)
		}
		strategy := "declared-recipe"
		if _, exists := bySubject[logical]; !exists {
			strategy = "ast-generic"
		}
		subjects = append(subjects, extractionSubject{Logical: logical, Before: plans[logical].Lines, After: extractionLines(source.data), Files: files, CreatedFiles: createdFiles, Strategy: strategy, Consumer: "function-extractor", Operation: operationBySubject[logical], Proof: proofBySubject[logical]})
	}
	return staged, subjects, unhandled, failures, nil
}
