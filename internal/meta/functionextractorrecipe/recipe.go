// Package recipe owns the generic extraction recipe authority.
package recipe

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Manifest struct {
	Schema  string             `json:"schema"`
	Recipes []ExtractionRecipe `json:"recipes"`
}

type ExtractionRecipe struct {
	Subject   string         `json:"subject"`
	Operation string         `json:"operation"`
	Edits     []TextEdit     `json:"edits"`
	Creates   []FileCreation `json:"creates,omitempty"`
}

type TextEdit struct {
	Path string   `json:"path"`
	Old  []string `json:"old_lines"`
	New  []string `json:"new_lines"`
}

type FileCreation struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

//go:embed recipes.json
var embedded []byte

func Load() ([]ExtractionRecipe, error) {
	var manifest Manifest
	if err := json.Unmarshal(embedded, &manifest); err != nil {
		return nil, err
	}
	if manifest.Schema != "gooo.function-extraction-recipes.v1" {
		return nil, fmt.Errorf("unsupported recipe manifest %q", manifest.Schema)
	}
	return manifest.Recipes, nil
}

func Index(recipes []ExtractionRecipe) (map[string]ExtractionRecipe, error) {
	bySubject := make(map[string]ExtractionRecipe, len(recipes))
	for _, recipe := range recipes {
		if _, exists := bySubject[recipe.Subject]; exists {
			return nil, fmt.Errorf("duplicate recipe %s", recipe.Subject)
		}
		bySubject[recipe.Subject] = recipe
	}
	return bySubject, nil
}
