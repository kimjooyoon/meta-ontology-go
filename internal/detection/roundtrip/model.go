// Package roundtrip reports semantic projection and generated-region locality
// violations at the compiler projection boundary.
//
// The detector is deliberately bound to the repository's semantic kernel. It
// does not define a second snapshot vocabulary that could drift from the
// authoritative IR.
package roundtrip

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// These aliases keep the detector API close to the semantic vocabulary while
// ensuring callers provide current semantic IR values.
type (
	Kind    = semantic.Kind
	Node    = semantic.Node
	Fact    = semantic.Fact
	IR      = semantic.IR
	ID      = semantic.ID
	FactKey = semantic.FactKey
)

const (
	Entity   = semantic.Entity
	Activity = semantic.Activity
	Agent    = semantic.Agent
)

// Observation is one complete projection witness. DSL and IR represent the
// authoritative source/lowering boundary; GoIR is the result lifted from
// generated Go. Before/AfterIR are optional snapshots for locality inference.
type Observation struct {
	DSL        semantic.IR
	IR         semantic.IR
	GoIR       semantic.IR
	BeforeIR   semantic.IR
	AfterIR    semantic.IR
	BeforeGo   []byte
	AfterGo    []byte
	AllowedIDs []semantic.ID
}

// LocalityInput supplies generated source and the identities allowed to move.
// Empty AllowedIDs means that every changed generated region is a violation.
type LocalityInput struct {
	Before     []byte
	After      []byte
	AllowedIDs []semantic.ID
}

// Delta is the deterministic semantic change between two valid IR snapshots.
// MetadataChanged covers semantic IR metadata that is not represented by a
// node or fact delta, such as package or namespace.
type Delta struct {
	AddedNodes      []semantic.Node
	RemovedNodes    []semantic.Node
	AddedFacts      []semantic.Fact
	RemovedFacts    []semantic.Fact
	TouchedIDs      []semantic.ID
	AffectedIDs     []semantic.ID
	MetadataChanged bool
}

// Equivalent reports semantic equivalence using the current kernel's
// validation and semantic hash rules. Presentation names, aliases, spans,
// candidate observations, and source formatting are not semantic meaning.
func Equivalent(left, right semantic.IR) bool {
	return semantic.CompareIR(left, right).SemanticEqual
}

// Fingerprint returns the kernel's stable semantic digest.
func Fingerprint(ir semantic.IR) string { return ir.StableHash() }
