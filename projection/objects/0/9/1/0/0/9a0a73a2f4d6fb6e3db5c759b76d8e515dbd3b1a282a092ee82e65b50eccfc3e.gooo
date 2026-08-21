package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func validateAnalyzeGeneratedSource(model generator.SemanticIR, authority semantic.IR, source []byte) error {
	text := string(source)
	if !strings.Contains(text, "//gooo:generated:") && !strings.Contains(text, "//gooo:slot:") {
		return nil
	}
	if _, err := generator.Generate(model, source); err != nil {
		return err
	}
	regions, slots, err := analyzeMarkers(source)
	if err != nil {
		return err
	}
	nodes := map[string]semantic.Kind{}
	for _, node := range authority.Graph.Nodes() {
		nodes[string(node.ID)] = node.Kind
	}
	seen := map[string]bool{}
	for _, region := range regions {
		if _, ok := nodes[region.id]; !ok {
			return fmt.Errorf("stale generated region identity %q", region.id)
		}
		if seen[region.id] {
			return fmt.Errorf("duplicate generated region identity %q", region.id)
		}
		seen[region.id] = true
	}
	for _, slot := range slots {
		valid := false
		for id, kind := range nodes {
			if kind == semantic.Activity && slot == id+"/implementation" {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("stale protected slot identity %q", slot)
		}
	}
	return nil
}
