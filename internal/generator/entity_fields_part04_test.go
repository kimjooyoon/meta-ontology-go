package generator

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"strings"
	"testing"
)

func TestEntityFieldsProfileBindingFailsClosed(t *testing.T) {
	ir := entityFieldsFixture()
	cases := []struct {
		name string
		edit func(*entityFieldsSupport)
		code string
	}{
		{name: "unknown state", edit: func(s *entityFieldsSupport) { s.State = "UNKNOWN" }, code: entityFieldsUnknownStateDiagnostic},
		{name: "unbound", edit: func(s *entityFieldsSupport) { s.Profile = syntax.EntityFieldsProfile{} }, code: entityFieldsUnboundProfileDiagnostic},
		{name: "profile ID", edit: func(s *entityFieldsSupport) { s.Profile.ID = "other" }, code: entityFieldsProfileMismatchDiagnostic},
		{name: "profile version", edit: func(s *entityFieldsSupport) { s.Profile.Version++ }, code: entityFieldsProfileMismatchDiagnostic},
		{name: "profile digest", edit: func(s *entityFieldsSupport) { s.Profile.Digest = "tampered" }, code: entityFieldsProfileDigestDiagnostic},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			support := supportedEntityFieldsForTest()
			testCase.edit(&support)
			result, err := New(Options{}).generateWithEntityFieldsSupport(ir, nil, support)
			if err == nil || !strings.Contains(err.Error(), testCase.code) || result.Source != nil || result.SourceMap.Mappings != nil {
				t.Fatalf("profile rejection was not transactional: result=%#v err=%v", result, err)
			}
		})
	}
}
func TestEntityFieldsFailuresDoNotMutateCallerInputsOrPreviousSource(t *testing.T) {
	ir := entityFieldsFixture()
	first := supportedEntityFieldsResult(t, ir, nil)
	previous := append([]byte(nil), first.Source...)
	changed := copyIR(ir)
	changed.Entities[0].Fields[0].Presence = "optional"
	beforeIR := copyIR(changed)
	result, err := New(Options{}).generateWithEntityFieldsSupport(changed, previous, supportedEntityFieldsForTest())
	if err == nil || result.Source != nil || !bytes.Equal(previous, first.Source) || !reflect.DeepEqual(changed, beforeIR) {
		t.Fatalf("rejected generation was not no-write: result=%#v err=%v", result, err)
	}
	corrupt := append(append([]byte(nil), first.Source...), []byte("\n//gooo:slot:start id=\"orphan\"\n")...)
	beforeCorrupt := append([]byte(nil), corrupt...)
	result, err = New(Options{}).generateWithEntityFieldsSupport(ir, corrupt, supportedEntityFieldsForTest())
	if err == nil || result.Source != nil || !bytes.Equal(corrupt, beforeCorrupt) {
		t.Fatalf("tampered previous source was not no-write: result=%#v err=%v", result, err)
	}
}
func TestFieldlessProductionProjectionKeepsDeferredMetadataAbsentAndDeterministic(t *testing.T) {
	ir := SemanticIR{Package: "fieldless", Entities: []Entity{{ID: "urn:gooo:entity:empty", Name: "Empty", GoName: "Empty"}}}
	first, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Source, second.Source) || !reflect.DeepEqual(first.SourceMap, second.SourceMap) || first.Metadata.EntityFields != nil || first.Metadata.SourceDigest != second.Metadata.SourceDigest || first.Metadata.SemanticIRDigest != second.Metadata.SemanticIRDigest || first.Metadata.SourceMapDigest != second.Metadata.SourceMapDigest {
		t.Fatalf("fieldless projection was not byte/hash deterministic: first=%#v second=%#v", first.Metadata, second.Metadata)
	}
	canonical, err := first.CanonicalJSON()
	if err != nil || bytes.Contains(canonical, []byte("entity_fields")) {
		t.Fatalf("fieldless canonical metadata changed: err=%v json=%s", err, canonical)
	}
}
