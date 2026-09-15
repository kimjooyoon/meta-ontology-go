package replay

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func structureOnly(input semantic.IR) (semantic.IR, error) {
	out := semantic.NewIR(input.Package, input.Namespace)
	for _, node := range input.Graph.Nodes() {
		if err := out.AddNode(node); err != nil {
			return semantic.IR{}, err
		}
	}
	return out.Normalized()
}
