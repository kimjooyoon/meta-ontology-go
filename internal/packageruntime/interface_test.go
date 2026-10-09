package packageruntime

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func interfaceFixture() Manifest {
	return Manifest{Schema: ManifestSchema, Entry: EntrySpec{PackagePath: "app", Activity: "Choose"}, Packages: []PackageSpec{
		{Path: "app", Name: "app", Imports: []string{"records"}, Sources: []Source{{Filename: "app.gooo", Content: `package app
namespace app
import "records"
activity Choose(Profile, Text, Profile) -> Profile computes "return input0"
`}}},
		{Path: "records", Name: "records", Sources: []Source{
			{Filename: "record.gooo", Content: `package records
namespace records
entity Profile id "example://profile" fields {
 field count id "example://profile/count" type integer required one
 field note id "example://profile/note" type string optional one
}
`},
			{Filename: "text.gooo", Content: "package records\nnamespace records\nentity Text id \"example://text\"\n"},
		}},
	}}
}

func TestDescribeUsesValidatedTypesAndOriginalDeclarations(t *testing.T) {
	manifest := interfaceFixture()
	before, _ := json.Marshal(manifest)
	image, err := Build(manifest)
	if err != nil {
		t.Fatal(err)
	}
	view, err := Describe(manifest)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) || view.ImageDigest != image.Digest || len(view.Packages) != 2 {
		t.Fatal("projection changed its source or build identity")
	}
	records, app := view.Packages[0], view.Packages[1]
	if records.Path != "records" || len(records.Declarations) != 2 || len(app.Declarations) != 1 {
		t.Fatal("projection lost declarations or exposed imported placeholder types", view)
	}
	profile := records.Declarations[0]
	if profile.Name != "Profile" || profile.ID != "example://profile" || profile.Source != "record.gooo" ||
		profile.Shape != "record" || len(profile.Fields) != 2 || profile.Fields[0].TypeID != "urn:gooo:type:integer" ||
		profile.Fields[0].Presence != "required" || profile.Fields[1].Presence != "optional" || profile.Fields[1].Cardinality != "one" {
		t.Fatal(profile)
	}
	if records.Declarations[1].Shape != "nominal" || len(records.Declarations[1].Fields) != 0 {
		t.Fatal(records)
	}
	activity := app.Declarations[0]
	if !reflect.DeepEqual(activity.Inputs, []string{"example://profile", "example://text", "example://profile"}) ||
		activity.Output != "example://profile" || activity.ID != "app://activity/choose" {
		t.Fatal(activity)
	}
	wantDigest := view.Digest
	view.Digest = ""
	if digestValue(view) != wantDigest {
		t.Fatal("interface digest does not bind its fields")
	}
}

func TestDescribeDeterministicOrderAndDetachedValues(t *testing.T) {
	manifest := interfaceFixture()
	a, err := Describe(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Packages[0], manifest.Packages[1] = manifest.Packages[1], manifest.Packages[0]
	sources := manifest.Packages[0].Sources
	sources[0], sources[1] = sources[1], sources[0]
	b, err := Describe(manifest)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal(err, "input order changed normalized interface")
	}
	a.Packages[0].Declarations[0].Fields[0].Name = "changed"
	a.Packages[1].Declarations[0].Inputs[0] = "changed"
	c, err := Describe(manifest)
	if err != nil || !reflect.DeepEqual(b, c) {
		t.Fatal(err, "caller mutation leaked into later projection")
	}
}

func TestDescribeExposesChangesUnderSameEntityIdentity(t *testing.T) {
	baseline, err := Describe(interfaceFixture())
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct{ name, from, to string }{
		{"field type", "type integer", "type boolean"},
		{"requiredness", "required one", "optional one"},
		{"field name", "field count", "field total"},
	} {
		t.Run(change.name, func(t *testing.T) {
			manifest := interfaceFixture()
			manifest.Packages[1].Sources[0].Content = strings.ReplaceAll(manifest.Packages[1].Sources[0].Content, change.from, change.to)
			view, err := Describe(manifest)
			if err != nil {
				t.Fatal(err)
			}
			old, next := baseline.Packages[0].Declarations[0], view.Packages[0].Declarations[0]
			if old.ID != next.ID || old.Fields[0].ID != next.Fields[0].ID || reflect.DeepEqual(old.Fields, next.Fields) || view.Digest == baseline.Digest {
				t.Fatal("shape change hidden by unchanged identity", next)
			}
		})
	}
}

func TestDescribeRejectsUnresolvedWorkspace(t *testing.T) {
	for _, change := range []func(*Manifest){
		func(m *Manifest) { m.Entry.Activity = "Missing" },
		func(m *Manifest) { m.Packages[0].Imports = []string{"missing"} },
		func(m *Manifest) {
			m.Packages[1].Sources[0].Content = strings.ReplaceAll(m.Packages[1].Sources[0].Content, "type integer", "type unknown")
		},
	} {
		manifest := interfaceFixture()
		change(&manifest)
		view, err := Describe(manifest)
		if err == nil || view.Schema != "" || len(view.Packages) != 0 {
			t.Fatal("partial graph represented as complete", err, view)
		}
	}
}

func TestDescribeRejectsFieldIdentityRepeatedAcrossFiles(t *testing.T) {
	manifest := interfaceFixture()
	manifest.Packages[1].Sources = append(manifest.Packages[1].Sources, Source{Filename: "other.gooo", Content: `package records
namespace records
entity Other id "example://other" fields {
 field other id "example://profile/count" type integer required one
}
`})
	view, err := Describe(manifest)
	if err == nil || !strings.Contains(err.Error(), "PACKAGE_INTERFACE_FIELD_INVALID") || view.Schema != "" {
		t.Fatal("cross-file field identity collision was hidden", err, view)
	}
}

func TestDescribeRetainsDeclaredEmptyRecord(t *testing.T) {
	manifest := interfaceFixture()
	manifest.Packages[1].Sources = append(manifest.Packages[1].Sources, Source{Filename: "empty.gooo", Content: `package records
namespace records
entity Empty id "example://empty" fields {}
`})
	view, err := Describe(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range view.Packages[0].Declarations {
		if item.Name == "Empty" {
			if item.Shape != "record" || len(item.Fields) != 0 {
				t.Fatal(item)
			}
			return
		}
	}
	t.Fatal("explicit empty record was omitted")
}
