// Package extractionplan prepares native source changes without committing files.
package extractionplan

import (
	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

type File struct {
	Name    string
	Data    []byte
	Mode    uint32
	Created bool
}

type Result struct {
	Files      map[string]File
	Recipes    map[string]recipeauthority.ExtractionRecipe
	Operations map[string][]string
	Evidence   map[string][]projectionextractor.StrategyEvidence
	Changed    map[string][]string
	Created    map[string][]string
	Failures   map[string]error
}

// Prepare is shared by the read-only planner and the committing bootstrap.
// Active recipes and the generic extractor keep their original precedence.
func Prepare(root string, residual []string, recipes []recipeauthority.ExtractionRecipe) (Result, error) {
	bySubject, err := recipeauthority.Index(recipes)
	if err != nil {
		return Result{}, err
	}
	buffers := make(map[string][]byte)
	created := make(map[string]bool)
	changedBySubject := make(map[string][]string)
	createdBySubject := make(map[string][]string)
	operationsBySubject := make(map[string][]string)
	evidenceBySubject := make(map[string][]projectionextractor.StrategyEvidence)
	failures := make(map[string]error)
	for _, logical := range residual {
		recipe, exists := bySubject[logical]
		if !exists {
			operations, evidence, err := stageGenericExtraction(root, logical, buffers, created, changedBySubject, createdBySubject)
			if err != nil {
				failures[logical] = err
				continue
			}
			operationsBySubject[logical] = operations
			evidenceBySubject[logical] = evidence
			continue
		}
		if err := stageRecipe(root, logical, recipe, buffers, created, changedBySubject, createdBySubject); err != nil {
			return Result{Failures: failures}, err
		}
	}
	staged, err := formatStaged(root, buffers, created)
	if err != nil {
		return Result{Failures: failures}, err
	}

	return Result{Files: staged, Recipes: bySubject, Operations: operationsBySubject,
		Evidence: evidenceBySubject, Changed: changedBySubject, Created: createdBySubject,
		Failures: failures}, nil
}
