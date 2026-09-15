package semanticdelta

// FormatVersion identifies the stable JSON and text interchange format.
const FormatVersion = "semanticdelta/v1"

// Format selects an interchange encoding.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Operation identifies whether a delta member was added or removed.
type Operation string

const (
	OperationAdd    Operation = "add"
	OperationRemove Operation = "remove"
)

// ChangeKind identifies the semantic item described by a violation.
type ChangeKind string

const (
	ChangeNode ChangeKind = "node"
	ChangeFact ChangeKind = "fact"
)

// Node is the identity-bearing part of a semantic IR node. Presentation and
// source evidence are intentionally absent from this boundary.
type Node struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// Fact is a directed semantic relation. Its endpoints are stable semantic
// identities, not display names or file paths.
type Fact struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

// Snapshot is the adapter-neutral semantic graph view used to compute a delta.
type Snapshot struct {
	Nodes []Node `json:"nodes,omitempty"`
	Facts []Fact `json:"facts,omitempty"`
}

// Delta is a presentation-insensitive semantic change set.
type Delta struct {
	AddedNodes   []Node `json:"addedNodes,omitempty"`
	RemovedNodes []Node `json:"removedNodes,omitempty"`
	AddedFacts   []Fact `json:"addedFacts,omitempty"`
	RemovedFacts []Fact `json:"removedFacts,omitempty"`
}

// Scope contains the semantic identities and predicates a change may touch.
// An empty list of predicates permits every predicate. Prefixes use ordinary
// string-prefix matching, which is useful for a namespace URI boundary.
type Scope struct {
	IDs        []string `json:"ids,omitempty"`
	Prefixes   []string `json:"prefixes,omitempty"`
	Predicates []string `json:"predicates,omitempty"`
}

// Request is the deterministic input accepted by Decode and the JSON/text
// encoders.
type Request struct {
	Version string `json:"version"`
	Allowed Scope  `json:"allowed"`
	Delta   Delta  `json:"delta"`
}
