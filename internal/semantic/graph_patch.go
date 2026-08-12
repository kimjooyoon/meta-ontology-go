package semantic

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	GraphPatchSchemaVersion                     = "gooo-graph-patch/v1"
	GraphPatchSetNodeField                      = "set-node-field"
	GraphPatchAddFact                           = "add-fact"
	PatchInvalidRequest       PatchConflictCode = "invalid_request"
	PatchStaleGraphHash       PatchConflictCode = "stale_graph_hash"
	PatchUnknownNode          PatchConflictCode = "unknown_node"
	PatchNodeHashMismatch     PatchConflictCode = "node_hash_mismatch"
	PatchUnknownField         PatchConflictCode = "unknown_field"
	PatchImmutableField       PatchConflictCode = "immutable_field"
	PatchFieldHashMismatch    PatchConflictCode = "field_hash_mismatch"
	PatchInvalidRelation      PatchConflictCode = "invalid_relation"
	PatchUnknownEndpoint      PatchConflictCode = "unknown_endpoint"
	PatchEndpointKindMismatch PatchConflictCode = "endpoint_kind_mismatch"
	PatchBaseTupleMismatch    PatchConflictCode = "base_tuple_mismatch"
	PatchIntentMissing        PatchConflictCode = "intent_missing"
	PatchLocalityMissing      PatchConflictCode = "locality_missing"
)

var ErrGraphPatch = errors.New("semantic graph patch rejected")

// PatchConflictCode identifies a deterministic, pre-write rejection reason.
type PatchConflictCode string

// GraphPatchBase is the trusted source/IR tuple observed by the validator.
type GraphPatchBase struct {
	SourceDigest string
	IRDigest     string
}

// GraphPatchRequest carries preconditions for one typed, not-yet-applied edit.
// This API validates only; it never changes graph, source, or store state.
type GraphPatchRequest struct {
	SchemaVersion        string
	Operation            string
	ExpectedGraphHash    string
	NodeID               ID
	ExpectedNodeHash     string
	Field                string
	ExpectedFieldHash    string
	Subject              ID
	Predicate            Relation
	Object               ID
	ExpectedSourceDigest string
	ExpectedIRDigest     string
	AllowedIntent        string
	Locality             string
}

// GraphPatchMutation is the in-memory payload applied after preconditions.
type GraphPatchMutation struct {
	Name    string
	Aliases []string
	Fact    *Fact
}

// GraphPatchConflict is a stable error code and human-readable detail.
type GraphPatchConflict struct {
	Code   PatchConflictCode
	Detail string
}

func (e GraphPatchConflict) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("%s: %s", ErrGraphPatch, e.Code)
	}
	return fmt.Sprintf("%s: %s: %s", ErrGraphPatch, e.Code, e.Detail)
}

func (e GraphPatchConflict) Unwrap() error { return ErrGraphPatch }

// NodeFieldHash returns a schema-bound digest for a mutable presentation field.
func NodeFieldHash(node Node, field string) (string, error) {
	normalized, err := node.Normalized()
	if err != nil {
		return "", patchConflict(PatchInvalidRequest, "node: "+err.Error())
	}
	var value strings.Builder
	value.WriteString("gooo-graph-node-field/v1\n")
	value.WriteString(field)
	value.WriteByte('\n')
	switch field {
	case "name":
		writeCanonicalField(&value, normalized.Name)
	case "aliases":
		for _, alias := range normalized.Aliases {
			writeCanonicalField(&value, alias)
		}
	default:
		return "", patchConflict(PatchUnknownField, field)
	}
	return StableHashString(value.String()), nil
}

// ValidatePatchPreconditions rejects stale or malformed edits before mutation.
func (g Graph) ValidatePatchPreconditions(base GraphPatchBase, request GraphPatchRequest) error {
	if request.SchemaVersion != GraphPatchSchemaVersion {
		return patchConflict(PatchInvalidRequest, "unsupported schema version")
	}
	if request.Operation != GraphPatchSetNodeField && request.Operation != GraphPatchAddFact {
		return patchConflict(PatchInvalidRequest, "unsupported operation")
	}
	if err := requireDigest(request.ExpectedGraphHash, "expected graph hash"); err != nil {
		return err
	}
	if request.ExpectedGraphHash != g.StableHash() {
		return patchConflict(PatchStaleGraphHash, "expected graph hash does not match")
	}
	if err := validatePatchBase(base, request); err != nil {
		return err
	}
	if strings.TrimSpace(request.AllowedIntent) == "" {
		return patchConflict(PatchIntentMissing, "allowed intent is required")
	}
	if strings.TrimSpace(request.Locality) == "" {
		return patchConflict(PatchLocalityMissing, "locality is required")
	}
	switch request.Operation {
	case GraphPatchSetNodeField:
		return g.validateNodeFieldPatch(request)
	case GraphPatchAddFact:
		return g.validateFactPatch(request)
	default:
		return patchConflict(PatchInvalidRequest, "unsupported operation")
	}
}

// ApplyGraphPatch validates and applies a typed mutation to a deep graph copy.
func (g Graph) ApplyGraphPatch(base GraphPatchBase, request GraphPatchRequest, mutation GraphPatchMutation) (Graph, error) {
	if err := g.Validate(); err != nil {
		return Graph{}, patchConflict(PatchInvalidRequest, "graph: "+err.Error())
	}
	if err := g.ValidatePatchPreconditions(base, request); err != nil {
		return Graph{}, err
	}
	clone := g.Clone()
	switch request.Operation {
	case GraphPatchSetNodeField:
		if err := clone.applyNodeMutation(request, mutation); err != nil {
			return Graph{}, err
		}
	case GraphPatchAddFact:
		if err := clone.applyFactMutation(request, mutation); err != nil {
			return Graph{}, err
		}
	default:
		return Graph{}, patchConflict(PatchInvalidRequest, "unsupported operation")
	}
	if err := clone.Validate(); err != nil {
		return Graph{}, patchConflict(PatchInvalidRequest, "resulting graph: "+err.Error())
	}
	return clone, nil
}

func (g *Graph) applyNodeMutation(request GraphPatchRequest, mutation GraphPatchMutation) error {
	id, err := ParseIdentity(request.NodeID.String())
	if err != nil {
		return patchConflict(PatchInvalidRequest, "node ID: "+err.Error())
	}
	node, ok := g.Node(id)
	if !ok {
		return patchConflict(PatchUnknownNode, id.String())
	}
	switch request.Field {
	case "name":
		node.Name = mutation.Name
	case "aliases":
		node.Aliases = append([]string(nil), mutation.Aliases...)
	default:
		return patchConflict(PatchInvalidRequest, "unsupported mutation field")
	}
	if err := g.AddNode(node); err != nil {
		return patchConflict(PatchInvalidRequest, "node mutation: "+err.Error())
	}
	return nil
}

func (g *Graph) applyFactMutation(request GraphPatchRequest, mutation GraphPatchMutation) error {
	if mutation.Fact == nil {
		return patchConflict(PatchInvalidRequest, "fact mutation is required")
	}
	fact, err := mutation.Fact.Normalized()
	if err != nil {
		return patchConflict(PatchInvalidRequest, "fact mutation: "+err.Error())
	}
	if fact.Status != FactDeterministic || fact.Subject != request.Subject ||
		fact.Predicate != request.Predicate || fact.Object != request.Object {
		return patchConflict(PatchInvalidRequest, "fact mutation does not match request")
	}
	if err := g.AddFact(fact); err != nil {
		return patchConflict(PatchInvalidRequest, "fact mutation: "+err.Error())
	}
	return nil
}

func validatePatchBase(base GraphPatchBase, request GraphPatchRequest) error {
	if err := requireDigest(request.ExpectedSourceDigest, "expected source digest"); err != nil {
		return err
	}
	if err := requireDigest(request.ExpectedIRDigest, "expected IR digest"); err != nil {
		return err
	}
	if err := requireDigest(base.SourceDigest, "source digest"); err != nil {
		return patchConflict(PatchBaseTupleMismatch, err.Error())
	}
	if err := requireDigest(base.IRDigest, "IR digest"); err != nil {
		return patchConflict(PatchBaseTupleMismatch, err.Error())
	}
	if base.SourceDigest != request.ExpectedSourceDigest || base.IRDigest != request.ExpectedIRDigest {
		return patchConflict(PatchBaseTupleMismatch, "source or IR digest does not match")
	}
	return nil
}

func (g Graph) validateNodeFieldPatch(request GraphPatchRequest) error {
	id, err := ParseIdentity(request.NodeID.String())
	if err != nil {
		return patchConflict(PatchInvalidRequest, "node ID: "+err.Error())
	}
	node, ok := g.Node(id)
	if !ok {
		return patchConflict(PatchUnknownNode, id.String())
	}
	if err := requireDigest(request.ExpectedNodeHash, "expected node hash"); err != nil {
		return err
	}
	if request.ExpectedNodeHash != node.StableHash() {
		return patchConflict(PatchNodeHashMismatch, id.String())
	}
	if request.Field != "name" && request.Field != "aliases" {
		if request.Field == "id" || request.Field == "kind" || request.Field == "namespace" {
			return patchConflict(PatchImmutableField, request.Field)
		}
		return patchConflict(PatchUnknownField, request.Field)
	}
	actual, err := NodeFieldHash(node, request.Field)
	if err != nil {
		return err
	}
	if err := requireDigest(request.ExpectedFieldHash, "expected field hash"); err != nil {
		return err
	}
	if request.ExpectedFieldHash != actual {
		return patchConflict(PatchFieldHashMismatch, request.Field)
	}
	return nil
}

func (g Graph) validateFactPatch(request GraphPatchRequest) error {
	subject, err := g.patchNode(request.Subject, "subject")
	if err != nil {
		return err
	}
	object, err := g.patchNode(request.Object, "object")
	if err != nil {
		return err
	}
	if !request.Predicate.Valid() {
		return patchConflict(PatchInvalidRelation, request.Predicate.String())
	}
	if err := request.Predicate.ValidateKinds(subject.Kind, object.Kind); err != nil {
		return patchConflict(PatchEndpointKindMismatch, err.Error())
	}
	return nil
}

func (g Graph) patchNode(raw ID, role string) (Node, error) {
	id, err := ParseIdentity(raw.String())
	if err != nil {
		return Node{}, patchConflict(PatchInvalidRequest, role+" ID: "+err.Error())
	}
	node, ok := g.Node(id)
	if !ok {
		return Node{}, patchConflict(PatchUnknownEndpoint, id.String())
	}
	return node, nil
}

func requireDigest(value, label string) error {
	if len(value) != 64 || value != strings.ToLower(value) {
		return patchConflict(PatchInvalidRequest, label+" must be lowercase SHA-256")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return patchConflict(PatchInvalidRequest, label+" must be lowercase SHA-256")
	}
	return nil
}

func patchConflict(code PatchConflictCode, detail string) error {
	return GraphPatchConflict{Code: code, Detail: detail}
}
