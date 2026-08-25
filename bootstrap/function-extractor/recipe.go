package main

import (
	"fmt"
	"os"
)

type recipeManifest struct {
	Schema  string             `json:"schema"`
	Recipes []extractionRecipe `json:"recipes"`
}

type extractionRecipe struct {
	Subject   string         `json:"subject"`
	Operation string         `json:"operation"`
	Edits     []textEdit     `json:"edits"`
	Creates   []fileCreation `json:"creates,omitempty"`
}

type textEdit struct {
	Path string   `json:"path"`
	Old  []string `json:"old_lines"`
	New  []string `json:"new_lines"`
}

type fileCreation struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

type stagedFile struct {
	name    string
	data    []byte
	mode    uint32
	created bool
}

func indexRecipes(recipes []extractionRecipe) (map[string]extractionRecipe, error) {
	bySubject := make(map[string]extractionRecipe, len(recipes))
	for _, recipe := range recipes {
		if _, exists := bySubject[recipe.Subject]; exists {
			return nil, fmt.Errorf("duplicate recipe %s", recipe.Subject)
		}
		bySubject[recipe.Subject] = recipe
	}
	return bySubject, nil
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
