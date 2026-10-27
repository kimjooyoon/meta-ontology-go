package workspaceexecution

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func TestPrepareKeepsRawActivityOriginsSeparateFromJSONStringImageDigests(t *testing.T) {
	manifest := importedActivityWorkspace()
	originals := map[string]string{}
	for pi := range manifest.Packages {
		pkg := &manifest.Packages[pi]
		for si := range pkg.Sources {
			source := &pkg.Sources[si]
			source.Content = "// 한글 <>& \"quoted\"\n" + source.Content
			originals[pkg.Path+"\x00"+source.Filename] = source.Content
		}
	}
	program, err := Prepare(manifest)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Image struct {
			Packages []struct {
				Path    string `json:"path"`
				Sources []struct {
					Filename string `json:"filename"`
					Digest   string `json:"source_digest"`
				} `json:"sources"`
			} `json:"packages"`
		} `json:"workspace_image"`
		Origins []struct {
			Input struct {
				Status   string `json:"status"`
				Package  string `json:"package_path"`
				Filename string `json:"filename"`
				Digest   string `json:"source_sha256"`
			} `json:"input_source"`
		} `json:"activity_origins"`
	}
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	imageDigests := map[string]string{}
	for _, pkg := range wire.Image.Packages {
		for _, source := range pkg.Sources {
			key := pkg.Path + "\x00" + source.Filename
			original, ok := originals[key]
			if !ok {
				t.Fatal("unmatched workspace source", key)
			}
			quoted, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("sha256:%x", sha256.Sum256(quoted))
			if source.Digest != want {
				t.Fatal("workspace image must hash the Go JSON string encoding", source)
			}
			imageDigests[key] = source.Digest
		}
	}
	if len(wire.Origins) != 2 {
		t.Fatal("reachable source origins missing", len(wire.Origins))
	}
	for _, origin := range wire.Origins {
		input := origin.Input
		key := input.Package + "\x00" + input.Filename
		original, ok := originals[key]
		if !ok || input.Status != "MAPPED" {
			t.Fatal("unmatched input source", input)
		}
		want := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(original)))
		if input.Digest != want || input.Digest == imageDigests[key] {
			t.Fatal("activity origin must hash raw full-file bytes independently of the image codec", input)
		}
	}
}
