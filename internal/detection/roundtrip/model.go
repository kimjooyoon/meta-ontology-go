// Package roundtrip reports semantic round-trip and generated-region locality
// violations at the compiler projection boundary.
//
// The package intentionally owns small adapter-neutral snapshots instead of
// importing a particular DSL, IR, analyzer, or generator implementation. A
// caller can therefore use it while those projections evolve independently.
package roundtrip

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Kind is the semantic role of a node.
type Kind string

const (
	Entity   Kind = "Entity"
	Activity Kind = "Activity"
	Agent    Kind = "Agent"
)

// Node is a semantic declaration. Name and namespace are presentation or
// lookup data; ID, Kind, and Attributes define its semantic meaning here.
type Node struct {
	ID         string
	Kind       Kind
	Name       string
	Namespace  string
	Attributes map[string]string
}

// Fact is a directed semantic relation between two stable identities.
type Fact struct {
	Subject    string
	Predicate  string
	Object     string
	Attributes map[string]string
}

// IR is the adapter-neutral semantic snapshot compared by the detector.
type IR struct {
	Package   string
	Namespace string
	Nodes     []Node
	Facts     []Fact
}

// Observation is one complete projection witness. DSL and IR represent the
// source/lowering boundary; GoIR is the result lifted from generated Go.
type Observation struct {
	DSL        IR
	IR         IR
	GoIR       IR
	BeforeIR   IR
	AfterIR    IR
	BeforeGo   []byte
	AfterGo    []byte
	AllowedIDs []string
}

// LocalityInput supplies generated source and the identities allowed to move.
// Empty AllowedIDs means that any changed generated region is a violation.
type LocalityInput struct {
	Before     []byte
	After      []byte
	AllowedIDs []string
}

// Delta is the semantic change between two IR snapshots.
type Delta struct {
	AddedNodes   []Node
	RemovedNodes []Node
	AddedFacts   []Fact
	RemovedFacts []Fact
	TouchedIDs   []string
	AffectedIDs  []string
}

// Validate rejects malformed snapshots before comparison can hide a defect.
func (ir IR) Validate() error {
	nodes := make(map[string]Kind, len(ir.Nodes))
	for _, node := range ir.Nodes {
		if err := validID(node.ID); err != nil {
			return fmt.Errorf("node %q: %w", node.ID, err)
		}
		if node.Kind != Entity && node.Kind != Activity && node.Kind != Agent {
			return fmt.Errorf("node %q has unknown kind %q", node.ID, node.Kind)
		}
		if previous, exists := nodes[node.ID]; exists {
			return fmt.Errorf("duplicate node ID %q (%s and %s)", node.ID, previous, node.Kind)
		}
		nodes[node.ID] = node.Kind
	}
	facts := make(map[string]struct{}, len(ir.Facts))
	for _, fact := range ir.Facts {
		if err := validID(fact.Subject); err != nil {
			return fmt.Errorf("fact subject %q: %w", fact.Subject, err)
		}
		if err := validID(fact.Object); err != nil {
			return fmt.Errorf("fact object %q: %w", fact.Object, err)
		}
		if strings.TrimSpace(fact.Predicate) == "" {
			return fmt.Errorf("fact %q -> %q has empty predicate", fact.Subject, fact.Object)
		}
		if _, exists := nodes[fact.Subject]; !exists {
			return fmt.Errorf("fact %s references unknown subject %q", fact.Predicate, fact.Subject)
		}
		if _, exists := nodes[fact.Object]; !exists {
			return fmt.Errorf("fact %s references unknown object %q", fact.Predicate, fact.Object)
		}
		key := factKey(fact)
		if _, exists := facts[key]; exists {
			return fmt.Errorf("duplicate fact %s", key)
		}
		facts[key] = struct{}{}
	}
	return nil
}

// Equivalent reports semantic equivalence while ignoring presentation fields.
func Equivalent(left, right IR) bool {
	if left.Validate() != nil || right.Validate() != nil {
		return false
	}
	left, right = left.normalized(), right.normalized()
	if len(left.Nodes) != len(right.Nodes) || len(left.Facts) != len(right.Facts) {
		return false
	}
	for index := range left.Nodes {
		if !sameNode(left.Nodes[index], right.Nodes[index]) {
			return false
		}
	}
	for index := range left.Facts {
		if !sameFact(left.Facts[index], right.Facts[index]) {
			return false
		}
	}
	return true
}

// Fingerprint returns a stable digest of semantic fields only.
func Fingerprint(ir IR) string {
	ir = ir.normalized()
	var canonical strings.Builder
	for _, node := range ir.Nodes {
		writePart(&canonical, node.ID)
		writePart(&canonical, string(node.Kind))
		writeMap(&canonical, node.Attributes)
	}
	canonical.WriteByte('|')
	for _, fact := range ir.Facts {
		writePart(&canonical, fact.Subject)
		writePart(&canonical, fact.Predicate)
		writePart(&canonical, fact.Object)
		writeMap(&canonical, fact.Attributes)
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	return hex.EncodeToString(digest[:])
}

func (ir IR) normalized() IR {
	result := IR{Package: ir.Package, Namespace: ir.Namespace}
	for _, node := range ir.Nodes {
		node.Attributes = cloneMap(node.Attributes)
		result.Nodes = append(result.Nodes, node)
	}
	for _, fact := range ir.Facts {
		fact.Attributes = cloneMap(fact.Attributes)
		result.Facts = append(result.Facts, fact)
	}
	sort.Slice(result.Nodes, func(i, j int) bool { return result.Nodes[i].ID < result.Nodes[j].ID })
	sort.Slice(result.Facts, func(i, j int) bool { return factKey(result.Facts[i]) < factKey(result.Facts[j]) })
	return result
}

func sameNode(left, right Node) bool {
	return left.ID == right.ID && left.Kind == right.Kind && mapsEqual(left.Attributes, right.Attributes)
}

func sameFact(left, right Fact) bool {
	return factKey(left) == factKey(right) && mapsEqual(left.Attributes, right.Attributes)
}

func factKey(fact Fact) string {
	return fact.Subject + "\x00" + fact.Predicate + "\x00" + fact.Object
}

func validID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("semantic ID is empty")
	}
	if strings.IndexFunc(id, func(r rune) bool { return r == '\n' || r == '\r' || r == '\t' || r == ' ' }) >= 0 {
		return fmt.Errorf("semantic ID contains whitespace")
	}
	return nil
}

func cloneMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func mapsEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func writePart(builder *strings.Builder, value string) {
	fmt.Fprintf(builder, "%d:", len(value))
	builder.WriteString(value)
}

func writeMap(builder *strings.Builder, values map[string]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteByte('{')
		writePart(builder, key)
		writePart(builder, values[key])
		builder.WriteByte('}')
	}
}
