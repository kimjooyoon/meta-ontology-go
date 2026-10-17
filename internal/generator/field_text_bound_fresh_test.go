package generator

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func TestProjectedTextBoundFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bound textbound.UTF8
		tag   string
	}{{"zero", textbound.New(0), `gooo_max_utf8_bytes:"0"`},
		{"expanded", textbound.New(1030), `gooo_max_utf8_bytes:"1030"`}} {
		t.Run(tc.name, func(t *testing.T) {
			ir := entityFieldsFixture()
			ir.Entities[0].Fields[0].TextBound = tc.bound
			result, err := GenerateWithEntityFieldsSupport(ir, nil, syntax.EntityFieldsV6Support())
			if err != nil || !strings.Contains(string(result.Source), tc.tag) {
				t.Fatal("Go projection lost the source bound", string(result.Source), err)
			}
			mappings := result.SourceMap.Lookup(ir.Entities[0].Fields[0].ID)
			if len(mappings) != 1 || mappings[0].TextBound != tc.bound {
				t.Fatal("source map lost the field constraint", mappings)
			}
			raw, err := json.Marshal(mappings[0])
			if err != nil || !bytes.Contains(raw, []byte(`"max_utf8_bytes":`)) {
				t.Fatal("source map did not publish the actual byte limit", string(raw), err)
			}
			if _, err := GenerateWithEntityFieldsSupport(ir, nil, syntax.EntityFieldsV5Support()); err == nil {
				t.Fatal("older projection accepted the new constraint")
			}
		})
	}
}

func TestUnboundedProjectionBytesPreservedFresh1905(t *testing.T) {
	ir := entityFieldsFixture()
	prior, err := GenerateWithEntityFieldsSupport(ir, nil, syntax.EntityFieldsV5Support())
	if err != nil {
		t.Fatal(err)
	}
	next, err := GenerateWithEntityFieldsSupport(ir, nil, syntax.EntityFieldsV6Support())
	if err != nil || !bytes.Equal(prior.Source, next.Source) {
		t.Fatal("undeclared bound changed structural Go bytes", err)
	}
	for _, mapping := range next.SourceMap.Mappings {
		raw, err := json.Marshal(mapping)
		if err != nil || bytes.Contains(raw, []byte("max_utf8_bytes")) {
			t.Fatal("undeclared constraint changed old mapping shape", string(raw), err)
		}
	}
}
