package semantic

import (
	"strings"
)

const CurrentIRVersion = "semantic-ir/v1"

// IR is the generalized semantic intermediate representation. Its Graph is
// parser-independent and can be populated by a DSL lowerer, a Go symbol
// lifter, or another authoritative view.
type IR struct {
	Version   string
	Package   string
	Namespace Namespace
	Graph     Graph
	evidence  map[ID]Evidence
}

// SemanticIR is an alias for callers that want the full name in APIs.
type SemanticIR = IR

func NewIR(packageName string, namespace Namespace) IR {
	return IR{
		Version:   CurrentIRVersion,
		Package:   strings.TrimSpace(packageName),
		Namespace: namespace,
		Graph:     NewGraph(),
		evidence:  make(map[ID]Evidence),
	}
}
func (ir *IR) AddNode(node Node) error {
	return ir.Graph.AddNode(node)
}
func (ir *IR) AddFact(fact Fact) error {
	return ir.Graph.AddFact(fact)
}
func (ir *IR) AddCandidate(fact Fact) error {
	return ir.Graph.AddCandidate(fact)
}
func (ir *IR) AddActivityContract(contract ActivityContract) error {
	return ir.Graph.AddActivityContract(contract)
}
func (ir IR) Validate() error {
	if err := validateIRVersion(ir.Version); err != nil {
		return err
	}
	if err := validatePackageName(ir.Package); err != nil {
		return err
	}
	if ir.Namespace != "" {
		if _, err := ParseNamespace(ir.Namespace.String()); err != nil {
			return err
		}
	}
	if err := ir.Graph.Validate(); err != nil {
		return err
	}
	return ir.validateEvidence()
}
