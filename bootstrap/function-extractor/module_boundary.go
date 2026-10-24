package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/moduleboundary"

func validateRecipeModule(root, logical string, recipe extractionRecipe) error {
	source, err := extractionPath(root, logical)
	if err != nil {
		return err
	}
	targets := make([]string, 0, len(recipe.Edits)+len(recipe.Creates))
	for _, edit := range recipe.Edits {
		target, err := extractionPath(root, edit.Path)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}
	for _, creation := range recipe.Creates {
		target, err := extractionPath(root, creation.Path)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}
	return moduleboundary.CheckWriteSet(root, source, targets...)
}
