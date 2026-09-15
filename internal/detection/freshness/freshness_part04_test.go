package freshness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsMalformedJSON(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		"null":         "null",
		"unknown":      "{\"sources\":[],\"unexpected\":true}",
		"duplicate":    "{\"sources\":[],\"sources\":[]}",
		"trailing":     "{\"sources\":[]} {}",
		"unterminated": "{\"sources\":[",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, name+".json")
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadManifest(path); err == nil {
				t.Fatalf("malformed manifest %q was accepted", data)
			}
		})
	}
}
func freshSnapshot(t *testing.T) (Snapshot, string) {
	t.Helper()
	root := t.TempDir()
	write := func(name, content string) string {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return name
	}
	sourcePath := write("main.gooo", "activity PayOrder")
	projectionPath := write("semantic.go", "generated projection")
	evidencePath := write("proof.json", "verification evidence")
	sourceDigest := HashBytes([]byte("activity PayOrder"))
	inputDigest, err := DigestInputs([]string{"dsl"}, map[string]string{"dsl": sourceDigest})
	if err != nil {
		t.Fatal(err)
	}
	return Snapshot{
		Root:    root,
		Sources: []Source{{ID: "dsl", Path: sourcePath}},
		Artifacts: []Artifact{{
			ID: "projection", Kind: KindProjection, Path: projectionPath,
			InputIDs: []string{"dsl"}, InputDigest: inputDigest,
			ContentDigest: HashBytes([]byte("generated projection")), EvidenceIDs: []string{"proof"},
		}},
		Evidence: []Evidence{{
			ID: "proof", Path: evidencePath, InputIDs: []string{"dsl"}, InputDigest: inputDigest,
			ContentDigest: HashBytes([]byte("verification evidence")),
			Provenance:    Provenance{ActivityID: "verify:generated", EntityID: "proof", UsedIDs: []string{"projection"}},
		}},
		RequiredArtifacts: []Requirement{{ID: "projection", Kind: KindProjection}},
		RequiredEvidence:  []Requirement{{ID: "proof", Kind: KindEvidence}},
	}, root
}
