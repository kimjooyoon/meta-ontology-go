package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"strings"
	"testing"
)

func TestEntityFieldsProjectionPreservesBidirIdentityOrderSpansAndProfile(t *testing.T) {
	ir, sourceModel := cliEntityFieldsFixture(t)
	if _, err := projectionIRFromBidirModel(ir, sourceModel); err == nil || !strings.Contains(err.Error(), "parse.entity-fields-deferred") {
		t.Fatalf("ordinary CLI projection activated fields while deferred: %v", err)
	}
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	projected, err := projectionIRFromBidirModelWithSupport(ir, sourceModel, support)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCLIEntityFieldsSupport(support); err != nil {
		t.Fatal(err)
	}
	wantProfile := syntax.EntityFieldsProfile{ID: syntax.EntityFieldsProfileID, Version: syntax.EntityFieldsProfileVersion, Digest: syntax.EntityFieldsProfileDigest}
	if support.Profile != wantProfile {
		t.Fatalf("profile = %#v, want %#v", support.Profile, wantProfile)
	}
	if len(projected.Entities) != 1 || len(projected.Entities[0].Fields) != 2 {
		t.Fatalf("projected fields = %#v", projected.Entities)
	}
	for index, got := range projected.Entities[0].Fields {
		want := sourceModel.Nodes[0].Fields[index]
		if got.ID != string(want.ID) || got.Parent != string(want.Parent) || got.Name != want.Name || got.TypeRefID != string(want.TypeRef.ID) || got.Presence != string(want.Presence) || got.Cardinality != string(want.Cardinality) {
			t.Fatalf("field %d lost identity/state: got=%#v want=%#v", index, got, want)
		}
		if got.Source != bidirGeneratorSpan(want.Span) || got.IDSpan != bidirGeneratorSpan(want.IDSpan) || got.NameSpan != bidirGeneratorSpan(want.NameSpan) || got.TypeRefSpan != bidirGeneratorSpan(want.TypeRefSpan) || got.PresenceSpan != bidirGeneratorSpan(want.PresenceSpan) || got.CardinalitySpan != bidirGeneratorSpan(want.CardinalitySpan) || got.NameSource != got.NameSpan {
			t.Fatalf("field %d lost exact source provenance: got=%#v want=%#v", index, got, want)
		}
	}
}
func TestEntityFieldsProjectionRejectsMissingAuthoritativeTypeRefIDWithoutWrites(t *testing.T) {
	ir, sourceModel := cliEntityFieldsFixture(t)
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	sourceModel.Nodes[0].Fields[0].TypeRef.ID = ""
	workspace := prepareDeferredCLIWorkspace(t)
	before := filesystemDigest(t, workspace.parent)
	var stdout, stderr bytes.Buffer
	projected, err := projectionIRFromBidirModelWithSupport(ir, sourceModel, support)
	wantErr := `entity "billing://entity/order" field 0: GOOO-EF-V1-INCOMPLETE-FIELD: field has no authoritative TypeRef.ID`
	if err == nil || err.Error() != wantErr || !reflect.DeepEqual(projected, generator.SemanticIR{}) {
		t.Fatalf("missing authoritative TypeRef.ID result = %#v err=%v, want exact rejection", projected, err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 || filesystemDigest(t, workspace.parent) != before {
		t.Fatalf("missing authoritative TypeRef.ID changed publication state: stdout=%q stderr=%q before=%s after=%s", stdout.String(), stderr.String(), before, filesystemDigest(t, workspace.parent))
	}
	assertDeferredCLIWorkspace(t, workspace)
}
