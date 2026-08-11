package semantic

import (
	"fmt"
	"strings"
	"unicode"
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
	if ir.Version == "" {
		return fmt.Errorf("%w: IR version is empty", ErrGraphInvalid)
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

func (ir IR) Normalized() (IR, error) {
	version := strings.TrimSpace(ir.Version)
	if version == "" {
		version = CurrentIRVersion
	}
	packageName := strings.TrimSpace(ir.Package)
	if err := validatePackageName(packageName); err != nil {
		return IR{}, err
	}
	namespace := ir.Namespace
	if namespace != "" {
		parsed, err := ParseNamespace(namespace.String())
		if err != nil {
			return IR{}, err
		}
		namespace = parsed
	}
	graph, err := ir.Graph.Normalized()
	if err != nil {
		return IR{}, err
	}
	out := IR{Version: version, Package: packageName, Namespace: namespace, Graph: graph, evidence: make(map[ID]Evidence)}
	for _, evidence := range ir.Evidence() {
		if err := out.AddEvidence(evidence); err != nil {
			return IR{}, err
		}
	}
	if err := out.validateEvidence(); err != nil {
		return IR{}, err
	}
	return out, nil
}

func (ir *IR) Normalize() error {
	normalized, err := ir.Normalized()
	if err != nil {
		return err
	}
	*ir = normalized
	return nil
}

func (ir IR) Canonical() string {
	version, packageName, namespace := canonicalIRMetadata(ir)
	var b strings.Builder
	b.WriteString("ir\t")
	b.WriteString(version)
	b.WriteByte('\t')
	b.WriteString(packageName)
	b.WriteByte('\t')
	b.WriteString(namespace)
	b.WriteByte('\n')
	b.WriteString(ir.Graph.Canonical())
	b.WriteString(ir.EvidenceCanonical())
	return b.String()
}

func (ir IR) SemanticCanonical() string {
	version, packageName, namespace := canonicalIRMetadata(ir)
	var b strings.Builder
	b.WriteString("ir\t")
	b.WriteString(version)
	b.WriteByte('\t')
	b.WriteString(packageName)
	b.WriteByte('\t')
	b.WriteString(namespace)
	b.WriteByte('\n')
	b.WriteString(ir.Graph.SemanticCanonical())
	return b.String()
}

func (ir IR) StableHash() string {
	return StableHashString(ir.SemanticCanonical())
}

func (ir IR) Hash() string {
	return ir.StableHash()
}

func canonicalIRMetadata(ir IR) (string, string, string) {
	version := strings.TrimSpace(ir.Version)
	if version == "" {
		version = CurrentIRVersion
	}
	packageName := strings.TrimSpace(ir.Package)
	namespace := strings.TrimSpace(ir.Namespace.String())
	if parsed, err := ParseNamespace(namespace); err == nil {
		namespace = parsed.String()
	}
	return version, packageName, namespace
}

func validatePackageName(packageName string) error {
	if packageName == "" {
		return nil
	}
	if strings.IndexFunc(packageName, unicode.IsSpace) >= 0 {
		return fmt.Errorf("%w: package name contains whitespace", ErrGraphInvalid)
	}
	return nil
}
