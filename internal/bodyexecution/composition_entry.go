package bodyexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

// A selected entry executes its explicit bind producers. Pure callees remain
// source declarations compiled inside a step, with arguments from the caller.
func compileCompositionEntry(document bidir.Document, model bidir.Model, entry string) (bidir.TypedPlan, error) {
	if entry == "" {
		return bidir.CompileTypedPlan(document)
	}
	needed := map[bidir.ID]bool{}
	identities := map[string]bidir.ID{}
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind {
			identities[node.Name] = node.ID
		}
	}
	for _, declaration := range document.Declarations {
		if declaration.Kind == bidir.ActivityKind && declaration.Name == entry {
			needed[identities[declaration.Name]] = true
		}
	}
	if len(needed) != 1 {
		return bidir.TypedPlan{}, fmt.Errorf("composition entry %q must name one activity", entry)
	}
	for changed := true; changed; {
		changed = false
		for _, edge := range document.BindingEdges {
			if needed[edge.TargetActivity] && !needed[edge.SourceActivity] {
				needed[edge.SourceActivity], changed = true, true
			}
		}
	}
	selected := document
	selected.Declarations = nil
	selected.BindingEdges = nil
	for _, declaration := range document.Declarations {
		if declaration.Kind == bidir.ActivityKind {
			declaration.ID = identities[declaration.Name]
		}
		if declaration.Kind != bidir.ActivityKind || needed[declaration.ID] {
			selected.Declarations = append(selected.Declarations, declaration)
		}
	}
	for _, edge := range document.BindingEdges {
		if needed[edge.TargetActivity] {
			selected.BindingEdges = append(selected.BindingEdges, edge)
		}
	}
	return bidir.CompileTypedPlan(selected)
}
