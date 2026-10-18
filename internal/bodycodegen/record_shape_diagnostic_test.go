package bodycodegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func recordShapeFixture(name string, count int) bidir.Node {
	id := bidir.ID("urn:gooo:record-diagnostic:" + name)
	node := bidir.Node{ID: id, Kind: bidir.EntityKind, Name: name}
	for index := range count {
		field := fmt.Sprintf("value%d", index)
		node.Fields = append(node.Fields, bidir.Field{ID: id + bidir.ID(":"+field), Parent: id,
			Name: field, TypeRef: bidir.TypeRef{ID: semantic.BuiltinIntegerTypeID},
			Presence: semantic.Required, Cardinality: semantic.One})
	}
	return node
}

func TestRecordShapeFieldDiagnosticKeepsSourceIdentity(t *testing.T) {
	for _, count := range []int{1, 16, 17, 32, 33} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			node := recordShapeFixture("Policy", count)
			records, err := RecordTypesFromModel(bidir.Model{Nodes: []bidir.Node{node}})
			if count <= 32 {
				if err != nil || len(records) != 1 || len(records[0].Fields) != count || records[0].ID != string(node.ID) {
					t.Fatal("valid record changed", records, err)
				}
				return
			}
			want := `record "Policy" (urn:gooo:record-diagnostic:Policy) declares 33 fields; body profile supports 1..32`
			if err == nil || err.Error() != want || len(records) != 0 {
				t.Fatal("field refusal lost its source or counts", err)
			}
		})
	}
}

func TestRecordShapeCountAndNameAreDistinctDiagnostics(t *testing.T) {
	var nodes []bidir.Node
	for i := range 16 {
		nodes = append(nodes, recordShapeFixture(fmt.Sprintf("R%d", i), 1))
	}
	if records, err := RecordTypesFromModel(bidir.Model{Nodes: nodes}); err != nil || len(records) != 16 {
		t.Fatal("sixteen records rejected", err)
	}
	nodes = append(nodes, recordShapeFixture("Overflow", 1))
	_, err := RecordTypesFromModel(bidir.Model{Nodes: nodes})
	if err == nil || err.Error() != `record "Overflow" (urn:gooo:record-diagnostic:Overflow) is record 17; body profile supports at most 16` {
		t.Fatal("record count refusal lost the triggering declaration", err)
	}
	node := recordShapeFixture("Policy", 1)
	node.Name = "not-an-identifier"
	_, err = RecordTypesFromModel(bidir.Model{Nodes: []bidir.Node{node}})
	if err == nil || err.Error() != `record "not-an-identifier" (urn:gooo:record-diagnostic:Policy) requires an identifier in body syntax` {
		t.Fatal("name refusal reported a size limit", err)
	}
}

func TestRecordShapeFieldDiagnosticReachesSourceCodegen(t *testing.T) {
	var source strings.Builder
	source.WriteString("package shape\nnamespace shape\nentity Policy id \"urn:gooo:shape:policy\" fields {\n")
	for i := range 33 {
		fmt.Fprintf(&source, "field f%d id \"urn:gooo:shape:f%d\" type integer required one\n", i, i)
	}
	source.WriteString("}\nactivity Echo(Policy) -> Policy computes `return input`\n")
	result, err := Generate("shape.gooo", []byte(source.String()), "Echo")
	if err == nil || !strings.Contains(err.Error(), `record "Policy" (urn:gooo:shape:policy) declares 33 fields; body profile supports 1..32`) || result.Source != "" {
		t.Fatal("source generation lost the precise refusal", err, result.Source)
	}
}
