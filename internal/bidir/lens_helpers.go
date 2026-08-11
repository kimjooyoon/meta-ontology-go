package bidir

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

func declarationIdentity(namespace string, declaration Declaration) (ID, error) {
	if declaration.ID != "" {
		if err := validateID(declaration.ID); err != nil {
			return "", fmt.Errorf("declaration %q: %w", declaration.Name, err)
		}
		return declaration.ID, nil
	}
	return ID(namespace + "://" + strings.ToLower(string(declaration.Kind)) + "/" + slug(declaration.Name)), nil
}

func resolveReference(reference Reference, namespace string, names map[string]ID, ids map[ID]struct{}) (ID, error) {
	if reference.ID != "" {
		if err := validateID(reference.ID); err != nil {
			return "", err
		}
		if _, exists := ids[reference.ID]; !exists {
			return "", fmt.Errorf("unknown semantic ID %q", reference.ID)
		}
		return reference.ID, nil
	}
	if strings.TrimSpace(reference.Name) == "" {
		return "", fmt.Errorf("reference has neither ID nor name")
	}
	refNamespace := reference.Namespace
	if refNamespace == "" {
		refNamespace = namespace
	}
	id, exists := names[referenceKey(refNamespace, reference.Name)]
	if !exists {
		return "", fmt.Errorf("unknown declaration %q in namespace %q", reference.Name, refNamespace)
	}
	return id, nil
}

func referenceKey(namespace, name string) string {
	return namespace + "\x00" + name
}

func appendUniqueRelation(relations []Relation, relation Relation) []Relation {
	key := relationKey(relation.Kind, relation.Source, relation.Target)
	for _, existing := range relations {
		if relationKey(existing.Kind, existing.Source, existing.Target) == key {
			return relations
		}
	}
	return append(relations, relation)
}

func appendCheckedRelation(relations []Relation, relation Relation) ([]Relation, error) {
	key := relationKey(relation.Kind, relation.Source, relation.Target)
	for _, existing := range relations {
		if relationKey(existing.Kind, existing.Source, existing.Target) != key {
			continue
		}
		if relationSemanticEqual(existing, relation) {
			return relations, nil
		}
		return nil, fmt.Errorf("duplicate relation %s with conflicting attributes", key)
	}
	return append(relations, relation), nil
}

func declarationFromNode(node Node, model Model) Declaration {
	declaration := Declaration{Kind: node.Kind, ID: node.ID, Name: node.Name, Attributes: cloneStringMap(node.Attributes), Span: node.Span}
	if declaration.Name == "" {
		declaration.Name = defaultName(node.ID)
	}
	if node.Kind != ActivityKind {
		return declaration
	}
	for _, relation := range model.Relations {
		if relation.Kind == PredicateUsed && relation.Source == node.ID {
			if target, exists := model.node(relation.Target); exists {
				declaration.Inputs = append(declaration.Inputs, Reference{ID: target.ID, Name: target.Name, Namespace: target.Namespace, Span: relation.Span})
			}
		}
		if relation.Kind == PredicateWasGeneratedBy && relation.Target == node.ID {
			if source, exists := model.node(relation.Source); exists {
				declaration.Outputs = append(declaration.Outputs, Reference{ID: source.ID, Name: source.Name, Namespace: source.Namespace, Span: relation.Span})
			}
		}
	}
	sort.SliceStable(declaration.Inputs, func(i, j int) bool {
		return referenceSourceOrderLess(declaration.Inputs[i], declaration.Inputs[j])
	})
	sort.Slice(declaration.Outputs, func(i, j int) bool { return declaration.Outputs[i].ID < declaration.Outputs[j].ID })
	return declaration
}

func referenceSourceOrderLess(left, right Reference) bool {
	leftValid, rightValid := left.Span.Valid(), right.Span.Valid()
	if leftValid != rightValid {
		return leftValid
	}
	if leftValid {
		if left.Span.File != right.Span.File {
			return left.Span.File < right.Span.File
		}
		if left.Span.Start != right.Span.Start {
			return left.Span.Start < right.Span.Start
		}
		if left.Span.StartLine != right.Span.StartLine {
			return left.Span.StartLine < right.Span.StartLine
		}
		if left.Span.StartColumn != right.Span.StartColumn {
			return left.Span.StartColumn < right.Span.StartColumn
		}
		if left.Span.End != right.Span.End {
			return left.Span.End < right.Span.End
		}
	}
	return left.ID < right.ID
}

func (m Model) node(id ID) (Node, bool) {
	for _, node := range m.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return Node{}, false
}

func implicitRelationKeys(declarations []Declaration, namespace string) map[string]struct{} {
	ids := make(map[string]ID, len(declarations))
	for _, declaration := range declarations {
		if id, err := declarationIdentity(namespace, declaration); err == nil {
			ids[referenceKey(namespace, declaration.Name)] = id
		}
	}
	result := make(map[string]struct{})
	for _, declaration := range declarations {
		if declaration.Kind != ActivityKind {
			continue
		}
		activityID, err := declarationIdentity(namespace, declaration)
		if err != nil {
			continue
		}
		for _, reference := range declaration.Inputs {
			id := reference.ID
			if id == "" {
				id = ids[referenceKey(namespace, reference.Name)]
			}
			if id != "" {
				result[relationKey(PredicateUsed, activityID, id)] = struct{}{}
			}
		}
		for _, reference := range declaration.Outputs {
			id := reference.ID
			if id == "" {
				id = ids[referenceKey(namespace, reference.Name)]
			}
			if id != "" {
				result[relationKey(PredicateWasGeneratedBy, id, activityID)] = struct{}{}
			}
		}
	}
	return result
}

func relationLess(left, right Relation) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	if left.Source != right.Source {
		return left.Source < right.Source
	}
	return left.Target < right.Target
}

func slug(value string) string {
	var builder strings.Builder
	lastDash := false
	var previous rune
	for _, current := range value {
		if unicode.IsUpper(current) && unicode.IsLower(previous) && builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
		}
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			builder.WriteRune(unicode.ToLower(current))
			lastDash = false
		} else if builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
		previous = current
	}
	return strings.Trim(builder.String(), "-")
}
