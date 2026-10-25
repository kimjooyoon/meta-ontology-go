package sourcebundle_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestPartitionRetainsExplicitProfileAndPhysicalFieldOrigins(t *testing.T) {
	for _, row := range []struct {
		file, boundary string
		support        syntax.EntityFieldsSupport
	}{
		{"../../examples/entity-fields-v1/main.gooo", "", syntax.EntityFieldsV1Support()},
		{"../../examples/agent-tool-inventory/main.gooo", "urn:gooo:agent-tool-inventory:inspect", syntax.EntityFieldsV5Support()},
	} {
		t.Run(row.support.Profile.ID, func(t *testing.T) {
			raw, err := os.ReadFile(row.file)
			if err != nil {
				t.Fatal(err)
			}
			inputs := []sourcebundle.Source{{Filename: "main.gooo", Content: string(raw)}}
			if _, err := sourcebundle.Lower(inputs); err == nil || !strings.Contains(err.Error(), "entity-fields-deferred") {
				t.Fatal("ordinary lowering silently activated entity fields", err)
			}
			before, err := sourcebundle.LowerWithEntityFieldsSupport(inputs, row.support)
			if err != nil {
				t.Fatal(err)
			}
			boundary := row.boundary
			if boundary == "" {
				boundary = before.Provenance.Declarations[1].ID
			}
			if _, err := sourcebundle.PreparePartition(inputs, "main.gooo", boundary, "activity.gooo"); err == nil {
				t.Fatal("ordinary partition silently activated entity fields")
			}
			plan, err := sourcebundle.PreparePartitionWithEntityFieldsSupport(inputs, "main.gooo", boundary,
				"activity.gooo", row.support)
			if err != nil || plan.EntityFieldsSupport == nil || *plan.EntityFieldsSupport != row.support {
				t.Fatal("explicit profile was lost", err)
			}
			if err := sourcebundle.ValidatePartition(inputs, plan); err != nil {
				t.Fatal(err)
			}
			if plan.Before.IR.SemanticCanonical() != plan.After.IR.SemanticCanonical() ||
				plan.Before.SourceDigest == plan.After.SourceDigest {
				t.Fatal("partition lost semantic equivalence or distinct physical identity")
			}
			for _, origin := range plan.After.Provenance.Declarations {
				if origin.ID == boundary && origin.Span.File != "activity.gooo" {
					t.Fatal("moved declaration retained its old physical address")
				}
			}
			plan.EntityFieldsSupport = nil
			if err := sourcebundle.ValidatePartition(inputs, plan); err == nil {
				t.Fatal("replay inferred a missing explicit profile")
			}
		})
	}
}

func TestPartitionRefusesUnknownProfilesEvenWithoutFields(t *testing.T) {
	inputs := []sourcebundle.Source{{Filename: "main.gooo", Content: "package p\nnamespace p\nentity A id \"urn:p:a\"\nentity B id \"urn:p:b\"\n"}}
	for _, change := range []func(*syntax.EntityFieldsSupport){
		func(p *syntax.EntityFieldsSupport) { p.State = "UNKNOWN" },
		func(p *syntax.EntityFieldsSupport) { p.Profile.ID = "unknown.profile" },
		func(p *syntax.EntityFieldsSupport) { p.Profile.Version++ },
		func(p *syntax.EntityFieldsSupport) { p.Profile.Digest = strings.Repeat("0", 64) },
	} {
		support := syntax.EntityFieldsV1Support()
		change(&support)
		if result, err := sourcebundle.LowerWithEntityFieldsSupport(inputs, support); err == nil ||
			!reflect.DeepEqual(result, sourcebundle.Bundle{}) {
			t.Fatal("invalid profile produced a lowered bundle", err)
		}
		if result, err := sourcebundle.PreparePartitionWithEntityFieldsSupport(inputs, "main.gooo", "urn:p:b", "b.gooo", support); err == nil ||
			!reflect.DeepEqual(result, sourcebundle.Partition{}) {
			t.Fatal("invalid profile produced a partition", err)
		}
		plan, err := sourcebundle.PreparePartition(inputs, "main.gooo", "urn:p:b", "b.gooo")
		if err != nil {
			t.Fatal(err)
		}
		plan.EntityFieldsSupport = &support
		if err := sourcebundle.ValidatePartition(inputs, plan); err == nil {
			t.Fatal("replay accepted a tampered profile")
		}
	}
}
