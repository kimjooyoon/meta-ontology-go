package projectionbundle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOwnedProjectionRetirementFresh1868(t *testing.T) {
	for _, name := range []string{"owned", "changed-obsolete", "foreign-family", "empty-next", "invalid-previous"} {
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			member := []byte("old second projection\n")
			prior := Manifest{Schema: Schema, SourceDigest: Digest([]byte("source1868")), Members: []Member{
				{Path: "policy.generated.go", Digest: Digest([]byte("first")), RegionIDs: []string{"urn:1868:first"}},
				{Path: "policy_part02.generated.go", Digest: Digest(member), RegionIDs: []string{"urn:1868:second"}},
			}}
			next := prior
			next.Members = append([]Member{}, prior.Members[:1]...)
			switch name {
			case "changed-obsolete":
				member = []byte("caller's changed member")
			case "foreign-family":
				prior.Members[0].Path = "foreign.generated.go"
			case "empty-next":
				next.Members = nil
			}
			raw, err := json.Marshal(prior)
			if err != nil {
				t.Fatal(err)
			}
			if name == "invalid-previous" {
				raw = []byte("broken manifest")
			}
			path := filepath.Join(directory, "policy.projection.json")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "policy_part02.generated.go"), member, 0600); err != nil {
				t.Fatal(err)
			}
			retired, err := ObsoleteMembers(path, next)
			if name == "owned" {
				if err != nil || len(retired) != 1 || retired[0] != "policy_part02.generated.go" {
					t.Fatalf("owned member retirement: %v %v", retired, err)
				}
			} else if err == nil {
				t.Fatal("conflicting member set accepted")
			}
			if _, err := os.Stat(filepath.Join(directory, "policy_part02.generated.go")); err != nil {
				t.Fatal("retirement planning changed a file")
			}
		})
	}
}

func TestOwnedProjectionAtomicWriteFresh1868(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projection.generated.go")
	if err := os.WriteFile(path, []byte("previous member"), 0600); err != nil {
		t.Fatal(err)
	}
	const after = "complete replacement member\n"
	if err := WriteAtomic(path, []byte(after)); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != after {
		t.Fatalf("atomic member replacement: %q %v", raw, err)
	}
	leftovers, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".gooo-projection-*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatal("temporary member was retained", leftovers, err)
	}
}
