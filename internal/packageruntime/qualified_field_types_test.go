package packageruntime

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestQualifiedScalarFieldsRetainFileLocalSourceReferences(t *testing.T) {
	manifest := Manifest{Schema: ManifestSchema, Entry: EntrySpec{PackagePath: "field171/app", Activity: "Echo"},
		Packages: []PackageSpec{
			{Path: "field171/counts", Name: "counts", Sources: []Source{{Filename: "counts.gooo", Content: `package counts
namespace field171_counts
entity Count id "urn:gooo:type:integer"
entity Label id "urn:gooo:type:string"
`}}},
			{Path: "field171/flags", Name: "flags", Sources: []Source{{Filename: "flags.gooo", Content: `package flags
namespace field171_flags
entity Count id "urn:gooo:type:boolean"
`}}},
			{Path: "field171/app", Name: "app", Imports: []string{"field171/counts", "field171/flags"}, Sources: []Source{
				{Filename: "ledger.gooo", Content: `package app
namespace field171_app
import data "field171/counts"
entity Ledger id "urn:field171:ledger" fields {
 field amount id "urn:field171:amount" type data.Count required one
 field note id "urn:field171:note" type data.Label optional one
}
activity Echo(Ledger) -> Ledger id "urn:field171:echo" computes "return input"
`},
				{Filename: "claim.gooo", Content: `package app
namespace field171_app
import data "field171/flags"
entity Claim id "urn:field171:claim" fields {
 field eligible id "urn:field171:eligible" type data.Count required one
}
`},
			}},
		}}
	before, _ := json.Marshal(manifest)
	result, err := Describe(manifest)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"urn:field171:amount": "urn:gooo:type:integer",
		"urn:field171:note": "urn:gooo:type:string", "urn:field171:eligible": "urn:gooo:type:boolean"}
	got, references := map[string]string{}, 0
	for _, pkg := range result.Packages {
		if pkg.Path != "field171/app" {
			continue
		}
		for _, declaration := range pkg.Declarations {
			for _, field := range declaration.Fields {
				got[field.ID] = field.TypeID
			}
		}
		for _, source := range pkg.Sources {
			for _, ref := range source.FieldTypes {
				references++
				if ref.TypeID != want[ref.FieldID] || !strings.HasPrefix(ref.Spelling, "data.") || ref.End <= ref.Start {
					t.Fatal("source reference lost its spelling, range or type ID", ref)
				}
				provider := "field171/counts"
				if ref.FieldID == "urn:field171:eligible" {
					provider = "field171/flags"
				}
				if ref.ImportPath != provider {
					t.Fatal("another file's import alias leaked", ref)
				}
				for _, original := range manifest.Packages[2].Sources {
					if original.Filename == source.Filename && original.Content[ref.Start:ref.End] != ref.Spelling {
						t.Fatal("reference span no longer points to original lookup", ref)
					}
				}
			}
		}
	}
	if !reflect.DeepEqual(got, want) || references != 3 {
		t.Fatal("canonical field types or original reference count differ", got, references)
	}
	formatted, diagnostics, err := syntax.FormatSourceWithEntityFieldsSupport("ledger.gooo",
		manifest.Packages[2].Sources[0].Content, syntax.EntityFieldsV4Support())
	if err != nil || diagnostics.HasErrors() || !strings.Contains(formatted, "type data.Count required one") {
		t.Fatal("qualified type spelling did not format", err, diagnostics)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("source normalization rewrote the supplied manifest")
	}
	observation, _ := json.Marshal(map[string]any{"references": references, "reference_total": 3,
		"canonical_field_types": got, "source_files": 4, "packages": 3, "input_unchanged": true,
		"formatter_preserved_qualified_spelling": true, "model_predictions": 0, "native_processes": 0})
	t.Log("GOOO_QUALIFIED_FIELD_SOURCE_OBSERVATION=" + string(observation))
}
