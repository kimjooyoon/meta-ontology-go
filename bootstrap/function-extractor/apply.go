package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractionplan"
)

type stagedFile = extractionplan.File

func stageExtractions(root string, plans map[string]planSubject, residual []string,
	recipes []extractionRecipe) (map[string]stagedFile, []extractionSubject, []string, []extractionFailureRecord, error) {
	prepared, err := extractionplan.Prepare(root, residual, recipes)
	unhandled := make([]string, 0)
	failures := make([]extractionFailureRecord, 0)
	for _, logical := range residual {
		if failure, exists := prepared.Failures[logical]; exists {
			unhandled = append(unhandled, logical)
			failures = append(failures, extractionFailure(logical, failure))
			fmt.Printf("function-extractor: unhandled=%s %v\n", logical, failure)
		}
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}
	subjects, err := extractionSubjects(plans, residual, prepared.Recipes, prepared.Operations,
		prepared.Evidence, prepared.Changed, prepared.Created, prepared.Files)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return prepared.Files, subjects, unhandled, failures, nil
}
