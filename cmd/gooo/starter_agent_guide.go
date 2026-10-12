package main

import (
	_ "embed"
	"maps"
)

//go:embed templates/AGENTS.md.template
var starterAgentGuide string

func withStarterGuide(files map[string]string) map[string]string {
	result := maps.Clone(files)
	result["AGENTS.md"] = starterAgentGuide
	return result
}
