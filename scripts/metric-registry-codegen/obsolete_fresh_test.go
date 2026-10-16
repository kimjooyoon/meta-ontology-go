package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func TestBundleObsoleteOwnershipFresh1663(t *testing.T) {
	for _, name := range []string{"owned", "modified", "foreign-output", "malformed-manifest"} {
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			content := []byte("previous generated member")
			old := projectionbundle.Manifest{Schema: projectionbundle.Schema,
				SourceDigest: projectionbundle.Digest([]byte("declaration")), Members: []projectionbundle.Member{
					{Path: "root.generated.go", Digest: projectionbundle.Digest(content), RegionIDs: []string{"fixture://root"}},
					{Path: "root_part02.generated.go", Digest: projectionbundle.Digest(content), RegionIDs: []string{"fixture://second"}},
				}}
			next := old
			next.Members = append([]projectionbundle.Member(nil), old.Members[:1]...)
			if name == "modified" {
				content = []byte("changed after generation")
			}
			if err := os.WriteFile(filepath.Join(directory, "root_part02.generated.go"), content, 0600); err != nil {
				t.Fatal(err)
			}
			if name == "foreign-output" {
				old.Members[0].Path = "other.generated.go"
			}
			raw, err := json.Marshal(old)
			if err != nil {
				t.Fatal(err)
			}
			if name == "malformed-manifest" {
				raw = []byte("{}")
			}
			path := filepath.Join(directory, "root.projection.json")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			obsolete, err := obsoleteMembers(path, next)
			if (err == nil) != (name == "owned") || name == "owned" &&
				(len(obsolete) != 1 || obsolete[0] != "root_part02.generated.go") {
				t.Fatal("obsolete ownership", obsolete, err)
			}
		})
	}
}

func TestBundleInvalidOutputFresh1667(t *testing.T) {
	directory := t.TempDir()
	projection := []byte("package fixture\n" +
		"//gooo:generated:start id=\"fixture://activity\" kind=\"activity\"\n" +
		"func value() int64 { return 1 }\n" +
		"//gooo:generated:end id=\"fixture://activity\" kind=\"activity\"\n")
	if err := writeBundle(filepath.Join(directory, "invalid[1].generated.go"), []byte("source"), projection); err == nil {
		t.Fatal("a glob-sensitive output name was accepted")
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatal("invalid plan wrote a partial output", err)
	}
}
