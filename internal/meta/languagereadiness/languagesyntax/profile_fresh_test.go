package languagesyntax

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func freshProfileSource1841(kind, presence, cardinality string) []byte {
	return []byte(fmt.Sprintf(`package profile1841
namespace profile1841
entity Summary id "urn:profile1841:summary" fields {
 field value id "urn:profile1841:value" type %s %s %s
}
activity Echo(Summary) -> Summary id "urn:profile1841:echo" computes "return input"
`, kind, presence, cardinality))
}

func TestExactEntityProfileFresh1841(t *testing.T) {
	for _, item := range []struct {
		name, kind, presence, cardinality, profile string
		accept                                     bool
	}{
		{"bool-deferred", "boolean", "required", "one", "", false},
		{"bool-v1", "boolean", "required", "one", syntax.EntityFieldsProfileID, false},
		{"bool-v2", "boolean", "required", "one", syntax.EntityFieldsV2ProfileID, true},
		{"integer-v2", "integer", "required", "one", syntax.EntityFieldsV2ProfileID, false},
		{"integer-v3", "integer", "required", "one", syntax.EntityFieldsV3ProfileID, true},
		{"optional-v3", "string", "optional", "one", syntax.EntityFieldsV3ProfileID, false},
		{"optional-v4", "string", "optional", "one", syntax.EntityFieldsV4ProfileID, true},
		{"many-v4", "string", "required", "many", syntax.EntityFieldsV4ProfileID, false},
		{"many-v5", "string", "required", "many", syntax.EntityFieldsV5ProfileID, true},
		{"unknown", "boolean", "required", "one", "future-profile", false},
	} {
		t.Run(item.name, func(t *testing.T) {
			files := fstest.MapFS{"program.gooo": {Data: freshProfileSource1841(item.kind, item.presence, item.cardinality)}}
			definition := CaseDefinition{ID: "profile1841", Path: "program.gooo", Kind: KindValid,
				EntityFieldsProfile: item.profile}
			got := ReplayCase(files, definition)
			if (got.ObservedDecision == DecisionPass) != item.accept {
				t.Fatalf("exact profile decision: %+v", got)
			}
			if item.accept && (!got.ASTReplayed || !got.ByteReplayed || !got.SemanticReplayed ||
				!got.GetPut || !got.PutGet || got.EntityFieldsProfile != item.profile || len(got.ProfileDigest) != 71) {
				t.Fatalf("profile origin or BX laws missing: %+v", got)
			}
		})
	}
}

func TestProfileSelectionBoundaryFresh1842(t *testing.T) {
	legacy := CaseDefinition{ID: "legacy1842", Path: "legacy.gooo", Kind: KindValid, EntityFields: true}
	raw, _ := json.Marshal(legacy)
	if strings.Contains(string(raw), "entity_fields_profile") {
		t.Fatal("new field changed the absent-profile wire shape")
	}
	files := fstest.MapFS{"legacy.gooo": {Data: freshProfileSource1841("string", "required", "one")}}
	got := ReplayCase(files, legacy)
	if got.ObservedDecision != DecisionPass || got.EntityFieldsProfile != "" || got.ProfileDigest != "" {
		t.Fatalf("legacy V1 flag behavior changed: %+v", got)
	}
	for _, flags := range []CaseDefinition{
		{EntityFields: true}, {OrderedCollections: true}, {ImplicitActivityPorts: true},
	} {
		flags.EntityFieldsProfile = syntax.EntityFieldsV2ProfileID
		if got := ReplayCase(nil, flags); got.ObservedDecision != DecisionClosed ||
			len(got.Diagnostics) != 1 || got.Diagnostics[0] != "registry.profile-mismatch" {
			t.Fatalf("ambiguous profile read source or selected a route: %+v", got)
		}
	}
}
