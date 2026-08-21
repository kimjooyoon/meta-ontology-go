package main

type recipeManifest struct {
	Schema  string             `json:"schema"`
	Recipes []extractionRecipe `json:"recipes"`
}

type extractionRecipe struct {
	Subject   string     `json:"subject"`
	Operation string     `json:"operation"`
	Edits     []textEdit `json:"edits"`
}

type textEdit struct {
	Path string   `json:"path"`
	Old  []string `json:"old_lines"`
	New  []string `json:"new_lines"`
}

type stagedFile struct {
	name string
	data []byte
	mode uint32
}
