package sourcebundle_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func TestBundlePartitionUsesTheWorkspaceBytePartition(t *testing.T) {
	raw, err := os.ReadFile("../../examples/shared-source-consumers/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Run", "계산"} {
		source := strings.ReplaceAll(string(raw), "Run", name)
		inputs := []sourcebundle.Source{{Filename: "shared.gooo", Content: source}}
		plan, err := sourcebundle.PreparePartition(inputs, "shared.gooo", "urn:shared:run", "run.gooo")
		if err != nil {
			t.Fatal(err)
		}
		manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
			Entry: packageruntime.EntrySpec{PackagePath: "shared", Activity: name},
			Packages: []packageruntime.PackageSpec{{Path: "shared", Name: "shared",
				Sources: []packageruntime.Source{{Filename: "shared.gooo", Content: source}}}}}
		workspace, err := packageruntime.SplitSource(manifest, "shared", "shared.gooo", "urn:shared:run", "run.gooo")
		if err != nil || !reflect.DeepEqual(plan.Files, workspace.Files) {
			t.Fatal("native and workspace consumers used different byte partitions", err)
		}
		if plan.Before.IR.StableHash() != plan.After.IR.StableHash() || plan.Before.SourceDigest == plan.After.SourceDigest ||
			len(plan.After.Provenance.Declarations) != 3 || inputs[0].Content != source {
			t.Fatal("source or semantic identity was lost")
		}
		replay, err := sourcebundle.PreparePartition(inputs, "shared.gooo", "urn:shared:run", "run.gooo")
		if err != nil || !reflect.DeepEqual(plan, replay) {
			t.Fatal("source partition did not replay", err)
		}
	}
}

func TestBundlePartitionIncludesPolicyDeclarationOrigins(t *testing.T) {
	raw, err := os.ReadFile("../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	cut := strings.Index(string(raw), "policy ")
	source := string(raw[:cut]) + "entity Input id \"urn:partition:input\"\n" + string(raw[cut:])
	plan, err := sourcebundle.PreparePartition([]sourcebundle.Source{{Filename: "policy.gooo", Content: source}},
		"policy.gooo", "gooo://meta-policy-compilation/policy/v4", "rules.gooo")
	if err != nil || len(plan.After.IR.Policies) != 1 || len(plan.After.Provenance.Declarations) != 2 {
		t.Fatal("policy partition failed", err)
	}
	for _, origin := range plan.After.Provenance.Declarations {
		if origin.ID == "gooo://meta-policy-compilation/policy/v4" && origin.Span.File != "rules.gooo" {
			t.Fatal("policy attributed to its previous file")
		}
	}
}

func TestBundlePartitionRefusesUnknownScopeAndOccupiedTargets(t *testing.T) {
	const source = "package p\nnamespace p\nentity A id \"urn:p:a\"\nentity B id \"urn:p:b\"\n"
	for _, row := range []struct{ file, id, into string }{
		{"main.gooo", "urn:p:a", "new.gooo"}, {"main.gooo", "B", "new.gooo"},
		{"absent.gooo", "urn:p:b", "new.gooo"}, {"main.gooo", "urn:p:b", "MAIN.gooo"},
		{"main.gooo", "urn:p:b", "../new.gooo"}, {"main.gooo", "urn:p:b", "new.go"},
	} {
		plan, err := sourcebundle.PreparePartition([]sourcebundle.Source{{Filename: "main.gooo", Content: source}},
			row.file, row.id, row.into)
		if err == nil || !reflect.DeepEqual(plan, sourcebundle.Partition{}) {
			t.Fatal("invalid request produced a partial plan", row, err)
		}
	}
}
