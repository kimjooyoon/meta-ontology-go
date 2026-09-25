package syntax

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSupportedEntityFieldsPreservesEmptyBlockAndFixedIDRename(t *testing.T) {
	support := supportedEntityFields()
	empty, diagnostics := ParseWithEntityFieldsSupport(`package p namespace n entity Empty id "urn:empty" fields {}`, support)
	if len(diagnostics) != 0 || empty == nil || !empty.Declarations[0].(*EntityDecl).FieldsPresent {
		t.Fatalf("empty block parse = %#v, %#v", empty, diagnostics)
	}
	emptyText, err := FormatWithEntityFieldsSupport(empty, support)
	if err != nil || !strings.Contains(emptyText, `entity Empty id "urn:empty" fields {}`) {
		t.Fatalf("empty block format = %q, %v", emptyText, err)
	}
	first, firstDiagnostics := ParseWithEntityFieldsSupport(`package p namespace n entity E id "urn:e" fields { field Old id "urn:f" type string required one }`, support)
	second, secondDiagnostics := ParseWithEntityFieldsSupport(`package p namespace n entity E id "urn:e" fields { field New id "urn:f" type string required one }`, support)
	if len(firstDiagnostics) != 0 || len(secondDiagnostics) != 0 || first.Declarations[0].(*EntityDecl).Fields[0].ID != second.Declarations[0].(*EntityDecl).Fields[0].ID {
		t.Fatalf("fixed-ID rename changed identity: %#v %#v", first, second)
	}
}
func TestExplicitEntityFieldsRemainDeferredOnOrdinaryPublicPath(t *testing.T) {
	source := `package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one }`
	file, diagnostics := ParseFile("fields.gooo", source)
	if file != nil || len(diagnostics) != 1 || diagnostics[0].Code != DiagEntityFieldsDeferred {
		t.Fatalf("ordinary parse = %#v, %#v", file, diagnostics)
	}
	formatted, formatDiagnostics, err := FormatSource("fields.gooo", source)
	if formatted != "" || !reflect.DeepEqual(formatDiagnostics, diagnostics) || err == nil {
		t.Fatalf("ordinary format = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
}
func TestEntityFieldsSupportMismatchFailsClosedForParserAndFormatter(t *testing.T) {
	source := `package p namespace n entity E id "urn:e" fields {}`
	profile := CurrentEntityFieldsSupport().Profile
	cases := []EntityFieldsSupport{
		{State: "UNKNOWN", Profile: profile},
		{State: EntityFieldsSupported, Profile: EntityFieldsProfile{ID: "wrong", Version: 1, Digest: profile.Digest}},
	}
	for _, support := range cases {
		file, diagnostics := ParseWithEntityFieldsSupport(source, support)
		if file != nil || len(diagnostics) != 1 || diagnostics[0].Code != DiagEntityFieldsConfiguration {
			t.Fatalf("mismatch parse = %#v, %#v", file, diagnostics)
		}
		if _, err := FormatWithEntityFieldsSupport(nil, support); !errors.Is(err, ErrEntityFieldsUnknownState) && !errors.Is(err, ErrEntityFieldsProfileMismatch) {
			t.Fatalf("mismatch format error = %v", err)
		}
	}
}
