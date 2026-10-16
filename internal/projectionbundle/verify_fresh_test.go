package projectionbundle

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

const bundlePartFixture = `package fixture
//gooo:generated:start id="fixture://activity" kind="activity"
func value() int64 { return 7 }
//gooo:generated:end id="fixture://activity" kind="activity"
`

func bundleFixture(t *testing.T) ([]byte, Manifest, fstest.MapFS) {
	t.Helper()
	source := []byte("source declaration")
	manifest := Manifest{Schema: Schema, SourceDigest: Digest(source), Members: []Member{
		{Path: "root.generated.go", Digest: Digest([]byte(bundlePartFixture)), RegionIDs: []string{"fixture://activity"}},
	}}
	files := fstest.MapFS{"root.generated.go": {Data: []byte(bundlePartFixture)}}
	return source, manifest, files
}

func TestProjectionBundleViewsFresh1659(t *testing.T) {
	for _, name := range []string{"exact", "changed-actual", "changed-compiled", "missing", "extra", "resealed-region", "wrong-source"} {
		t.Run(name, func(t *testing.T) {
			source, manifest, compiled := bundleFixture(t)
			_, _, actual := bundleFixture(t)
			switch name {
			case "changed-actual":
				actual["root.generated.go"].Data = []byte("changed")
			case "changed-compiled":
				compiled["root.generated.go"].Data = []byte("changed")
			case "missing":
				delete(actual, "root.generated.go")
			case "extra":
				actual["root_part02.generated.go"] = &fstest.MapFile{Data: []byte(bundlePartFixture)}
			case "resealed-region":
				manifest.Members[0].RegionIDs[0] = "fixture://different"
			case "wrong-source":
				source = []byte("different declaration")
			}
			raw, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := Verify(source, raw, compiled, actual); (err == nil) != (name == "exact") {
				t.Fatalf("bundle view verification: %v", err)
			}
		})
	}
}

func TestProjectionManifestRefusalsFresh1659(t *testing.T) {
	for _, name := range []string{"path-traversal", "glob-path", "duplicate-file", "duplicate-region", "trailing-object"} {
		t.Run(name, func(t *testing.T) {
			_, manifest, _ := bundleFixture(t)
			switch name {
			case "path-traversal":
				manifest.Members[0].Path = "../other.generated.go"
			case "glob-path":
				manifest.Members[0].Path = "*.generated.go"
			case "duplicate-file":
				manifest.Members = append(manifest.Members, manifest.Members[0])
			case "duplicate-region":
				manifest.Members[0].RegionIDs = append(manifest.Members[0].RegionIDs, "fixture://activity")
			}
			raw, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if name == "trailing-object" {
				raw = append(raw, []byte("{}")...)
			}
			if _, err := Decode(raw); err == nil {
				t.Fatal("invalid or ambiguous manifest accepted")
			}
		})
	}
}
