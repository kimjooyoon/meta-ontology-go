package bidir

import (
	"fmt"
	"sort"
)

const (
	PutSourceInvalid     = "put.source-invalid"
	PutModelInvalid      = "put.model-invalid"
	PutProvenanceMissing = "put.provenance-missing"
	PutWriteConflict     = "put.write-conflict"
)

// PutError reports a rejected write. NoWrite is true for every PutError: the
// returned document is the original source view, never a partially built one.
type PutError struct {
	Code    string
	NoWrite bool
	Err     error
}

func (e *PutError) Error() string {
	if e == nil {
		return "bidir put failed"
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *PutError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Put writes an updated model into a document while preserving source order
// for surviving declarations and explicit relations.
func Put(document Document, updated Model) (Document, error) {
	source, err := Get(document)
	if err != nil {
		return document, putError(PutSourceInvalid, err)
	}
	updated = updated.Normalized()
	if err := updated.Validate(); err != nil {
		return document, putError(PutModelInvalid, err)
	}
	if err := validatePutProvenance(source, updated); err != nil {
		return document, putError(PutProvenanceMissing, err)
	}
	result := Document{Package: document.Package, Namespace: document.Namespace}
	if result.Package == "" {
		result.Package = updated.Package
	}
	if result.Namespace == "" {
		result.Namespace = updated.Namespace
	}
	if result.Namespace == "" {
		result.Namespace = "gooo"
	}
	nodes := nodeMap(updated.Nodes)
	declarationIDs, err := appendSurvivingDeclarations(&result, document.Declarations, nodes, updated)
	if err != nil {
		return document, putError(PutWriteConflict, err)
	}
	appendNewDeclarations(&result, updated, nodes, declarationIDs)
	appendUpdatedRelations(&result, document.Relations, updated)
	return result, nil
}

func putError(code string, err error) error {
	return &PutError{Code: code, NoWrite: true, Err: err}
}

func validatePutProvenance(source, updated Model) error {
	delta := Diff(source, updated)
	for _, node := range append(delta.AddedNodes, delta.RemovedNodes...) {
		if !node.Span.Valid() {
			return fmt.Errorf("node %q semantic change has no source span", node.ID)
		}
	}
	for _, relation := range append(delta.AddedRelations, delta.RemovedRelations...) {
		if !relation.Span.Valid() {
			return fmt.Errorf("relation %s %q -> %q semantic change has no source span", relation.Kind, relation.Source, relation.Target)
		}
	}
	return nil
}

func appendSurvivingDeclarations(result *Document, declarations []Declaration, nodes map[ID]Node, updated Model) (map[ID]struct{}, error) {
	ids := make(map[ID]struct{}, len(declarations))
	for _, declaration := range declarations {
		id, err := declarationIdentity(result.Namespace, declaration)
		if err != nil {
			return nil, err
		}
		if _, exists := ids[id]; exists {
			return nil, fmt.Errorf("duplicate source declaration ID %q", id)
		}
		ids[id] = struct{}{}
		if node, exists := nodes[id]; exists {
			result.Declarations = append(result.Declarations, declarationFromNode(node, updated))
		}
	}
	return ids, nil
}

func appendNewDeclarations(result *Document, updated Model, nodes map[ID]Node, existing map[ID]struct{}) {
	ids := make([]ID, 0, len(updated.Nodes))
	for _, node := range updated.Nodes {
		if _, exists := existing[node.ID]; !exists {
			ids = append(ids, node.ID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		result.Declarations = append(result.Declarations, declarationFromNode(nodes[id], updated))
	}
}

func appendUpdatedRelations(result *Document, original []Relation, updated Model) {
	implicit := implicitRelationKeys(result.Declarations, result.Namespace)
	relations := make(map[string]Relation, len(updated.Relations))
	for _, relation := range updated.Relations {
		relation = relation.normalized()
		key := relationKey(relation.Kind, relation.Source, relation.Target)
		if _, representable := implicit[key]; representable && len(relation.Attributes) == 0 {
			continue
		}
		relations[key] = relation
	}
	seen := make(map[string]struct{}, len(relations))
	for _, relation := range original {
		key := relationKey(relation.Kind, relation.Source, relation.Target)
		if next, exists := relations[key]; exists {
			result.Relations = append(result.Relations, next)
			seen[key] = struct{}{}
		}
	}
	remaining := make([]Relation, 0, len(relations))
	for key, relation := range relations {
		if _, exists := seen[key]; !exists {
			remaining = append(remaining, relation)
		}
	}
	sort.Slice(remaining, func(i, j int) bool { return relationLess(remaining[i], remaining[j]) })
	result.Relations = append(result.Relations, remaining...)
}

// CheckGetPut verifies that a generic source view survives a write-back.
func CheckGetPut(document Document) error {
	model, err := Get(document)
	if err != nil {
		return err
	}
	written, err := Put(document, model)
	if err != nil {
		return err
	}
	observed, err := Get(written)
	if err != nil {
		return err
	}
	if !SemanticEquivalent(model, observed) {
		return fmt.Errorf("Get-Put violated: semantic model changed after write-back")
	}
	return nil
}

// CheckPutGet verifies that an accepted model is visible after write-back.
func CheckPutGet(document Document, model Model) error {
	written, err := Put(document, model)
	if err != nil {
		return err
	}
	observed, err := Get(written)
	if err != nil {
		return err
	}
	if !SemanticEquivalent(model, observed) {
		return fmt.Errorf("Put-Get violated: semantic model changed after write-back")
	}
	return nil
}
