package packageruntime

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQualifiedFieldProfilesRejectMissingAliasesAndNestedTypes(t *testing.T) {
	makeManifest := func(imported bool) Manifest {
		spelling, entity := "missing.Value", `entity Value id "urn:gooo:type:integer"`
		if imported {
			spelling, entity = "types.Value", `entity Value id "urn:field204:nested"`
		}
		return Manifest{Schema: ManifestSchema, Entry: EntrySpec{PackagePath: "field204/app", Activity: "Echo"},
			Packages: []PackageSpec{
				{Path: "field204/types", Name: "types", Sources: []Source{{Filename: "types.gooo",
					Content: "package types\nnamespace field204_types\n" + entity + "\n"}}},
				{Path: "field204/app", Name: "app", Imports: []string{"field204/types"}, Sources: []Source{{Filename: "app.gooo",
					Content: "package app\nnamespace field204_app\nimport types \"field204/types\"\n" +
						"entity Box id \"urn:field204:box\" fields {\n field value id \"urn:field204:value\" type " + spelling + " required one\n}\n" +
						"activity Echo(Box) -> Box id \"urn:field204:echo\" computes \"return input\"\n"}}},
			}}
	}
	for _, imported := range []bool{false, true} {
		manifest := makeManifest(imported)
		before, _ := json.Marshal(manifest)
		result, err := Build(manifest)
		want := "PACKAGE_FIELD_TYPE_IMPORTED_INVALID"
		if imported {
			want = "PACKAGE_FIELD_TYPE_PROFILE_UNSUPPORTED"
		}
		if err == nil || !strings.Contains(err.Error(), want) || result.Digest != "" {
			t.Fatal("invalid field type yielded a partial accepted image", result, err)
		}
		after, _ := json.Marshal(manifest)
		if string(before) != string(after) {
			t.Fatal("failed field resolution rewrote source")
		}
	}
	t.Log(`GOOO_QUALIFIED_FIELD_PROFILE_OBSERVATION={"rejected_profiles":2,"accepted_partial_images":0,"source_unchanged":true,"model_predictions":0,"native_processes":0}`)
}
