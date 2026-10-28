package semantic

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestRecordReferencesShareNominalGraphAndExactFailureIdentity(t *testing.T) {
	makeNode := func(id, target ID) Node {
		return Node{ID: id, Kind: Entity, Fields: []Field{{ID: id + ":next", Parent: id,
			Name: "next", TypeRef: TypeRef{ID: target}, Presence: Required, Cardinality: One}}}
	}
	a, b, c := makeNode("urn:a", "urn:b"), makeNode("urn:b", "urn:c"), makeNode("urn:c", BuiltinIntegerTypeID)
	for _, test := range []struct {
		name   string
		nodes  []Node
		depth  int
		code   string
		field  ID
		target ID
	}{
		{name: "forward references", nodes: []Node{a, b, c}, depth: 3},
		{name: "closing edge", nodes: []Node{a, b, c}, depth: 2, code: "record-reference-depth", field: "urn:b:next", target: "urn:c"},
		{name: "cached child height", nodes: []Node{c, b, a}, depth: 2, code: "record-reference-depth", field: "urn:a:next", target: "urn:b"},
		{name: "missing target", nodes: []Node{a}, depth: 3, code: "record-reference-target", field: "urn:a:next", target: "urn:b"},
		{name: "empty target", nodes: []Node{makeNode("urn:a", "")}, depth: 3, code: "record-reference-target", field: "urn:a:next"},
		{name: "cycle", nodes: []Node{a, makeNode("urn:b", "urn:a")}, depth: 3, code: "record-reference-cycle", field: "urn:b:next", target: "urn:a"},
		{name: "duplicate identity", nodes: []Node{c, c}, depth: 3, code: "record-reference-definition", field: "urn:c", target: "urn:c"},
	} {
		t.Run(test.name, func(t *testing.T) {
			before, _ := json.Marshal(test.nodes)
			err := ValidateRecordReferences(test.nodes, test.depth)
			if test.code == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var issues *ValidationErrors
				if !errors.As(err, &issues) || len(issues.Issues) != 1 || issues.Issues[0] != (ValidationIssue{
					Code: test.code, Message: issues.Issues[0].Message, Subject: test.field, Object: test.target,
				}) {
					t.Fatal("nominal failure lost its exact source identities", err)
				}
			}
			after, _ := json.Marshal(test.nodes)
			if !bytes.Equal(before, after) {
				t.Fatal("graph validation changed its input")
			}
		})
	}
}
