package packageruntime

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQualifiedFieldSourceProfilesDeclineUnresolvedAndNominalMembers(t *testing.T) {
	makeManifest := func(imported bool) Manifest {
		spelling, entity := "missing.Value", `entity Value id "urn:gooo:type:integer"`
		if imported {
			spelling, entity = "types.Value", `entity Value id "urn:field248:nested"`
		}
		return Manifest{Schema: ManifestSchema, Entry: EntrySpec{PackagePath: "field248/app", Activity: "Echo"},
			Packages: []PackageSpec{
				{Path: "field248/types", Name: "types", Sources: []Source{{Filename: "types.gooo",
					Content: "package types\nnamespace field248_types\n" + entity + "\n"}}},
				{Path: "field248/app", Name: "app", Imports: []string{"field248/types"}, Sources: []Source{{Filename: "app.gooo",
					Content: "package app\nnamespace field248_app\nimport types \"field248/types\"\n" +
						"entity Box id \"urn:field248:box\" fields {\n field value id \"urn:field248:value\" type " + spelling + " required one\n}\n" +
						"activity Echo(Box) -> Box id \"urn:field248:echo\" computes \"return input\"\n"}}},
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
	t.Log(`GOOO_QUALIFIED_FIELD_DECLINE_OBSERVATION={"rejected_profiles":2,"accepted_partial_images":0,"source_unchanged":true,"model_predictions":0,"native_processes":0}`)
}
