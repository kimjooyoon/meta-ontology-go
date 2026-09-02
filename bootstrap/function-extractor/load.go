package main

import (
	"encoding/json"
	"fmt"
	"os"

	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
)

type extractionRecipe = recipeauthority.ExtractionRecipe
type textEdit = recipeauthority.TextEdit
type fileCreation = recipeauthority.FileCreation

func decodeJSONFile(name string, target any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func loadExtractionInputs(planName, densityName,
	expected string, fixedPoint bool) (map[string]planSubject, []string, error) {
	var plan splitPlan
	if err := decodeJSONFile(planName, &plan); err != nil {
		return nil, nil, err
	}
	var density densityReport
	if err := decodeJSONFile(densityName, &density); err != nil {
		return nil, nil, err
	}
	if plan.Schema != "gooo.logical-split-plan.v1" ||
		density.Schema != "gooo.line-density-rewrite.v1" {
		return nil, nil, fmt.Errorf("unsupported extraction evidence")
	}
	if plan.SourceSHA != expected || density.SourceSHA != expected {
		return nil, nil, fmt.Errorf("extraction evidence is not bound to %s", expected)
	}
	plans := make(map[string]planSubject, len(plan.Subjects))
	for _, subject := range plan.Subjects {
		if _, exists := plans[subject.Logical]; exists {
			return nil, nil, fmt.Errorf("duplicate split subject %s", subject.Logical)
		}
		plans[subject.Logical] = subject
	}
	residual := make([]string, 0)
	seen := make(map[string]bool)
	for _, subject := range density.Subjects {
		if subject.Status != "blocked" {
			continue
		}
		if seen[subject.Logical] {
			return nil, nil, fmt.Errorf("duplicate density subject %s", subject.Logical)
		}
		if _, exists := plans[subject.Logical]; !exists {
			return nil, nil, fmt.Errorf("density subject lacks split evidence: %s", subject.Logical)
		}
		seen[subject.Logical] = true
		residual = append(residual, subject.Logical)
	}
	if fixedPoint {
		for _, subject := range plan.Subjects {
			switch subject.Reason {
			case "no-movable-declaration", "fixed-declaration-capacity", "movable-declaration-capacity":
				if seen[subject.Logical] {
					continue
				}
				seen[subject.Logical] = true
				residual = append(residual, subject.Logical)
			}
		}
	}
	return plans, residual, nil
}

func loadRecipes() ([]extractionRecipe, error) {
	return recipeauthority.Load()
}
