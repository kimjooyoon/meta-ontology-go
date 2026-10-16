package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestProfileCLIProtocolFresh1854(t *testing.T) {
	root, requestPath, afterPath, directory, originals := profileCLIFixture1854(t)
	arguments := []string{"--root", root, "--request", requestPath, "--source-after", afterPath}
	var observed bytes.Buffer
	if err := runProfileMigration(append(arguments, "--inspect"), &observed, io.Discard); err != nil {
		t.Fatal(err)
	}
	var pins map[string]json.RawMessage
	if err := json.Unmarshal(observed.Bytes(), &pins); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	var request map[string]json.RawMessage
	if err := json.Unmarshal(raw, &request); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"snapshot_digest", "source_digest", "replacement_digest", "toolchain", "execution_identity"} {
		request[key] = pins[key]
	}
	raw, err = json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	profileCLIWrite1854(t, requestPath, raw)
	observed.Reset()
	if err := runProfileMigration(append(arguments, "--output", directory), &observed, io.Discard); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(directory, "candidate.json"))
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := syntaxregistration.DecodeProfileMigrationCandidate(raw)
	if err != nil {
		t.Fatal(err)
	}
	checkProfileCLIBytes1854(t, root, directory, originals, candidate)
	observed.Reset()
	if err := runProfileMigration(append(arguments, "--verify-candidate", filepath.Join(directory, "candidate.json")),
		&observed, io.Discard); err != nil || !bytes.Contains(observed.Bytes(), []byte("EXACT_SOURCE_CANDIDATE")) {
		t.Fatalf("CLI consumer failed: %v %s", err, observed.String())
	}
}

func checkProfileCLIBytes1854(t *testing.T, root, directory string, originals map[string][]byte,
	candidate syntaxregistration.ProfileMigrationCandidate) {
	t.Helper()
	for path, want := range originals {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || !bytes.Equal(raw, want) {
			t.Fatalf("input changed: %s: %v", path, err)
		}
	}
	for _, member := range candidate.Members {
		for _, item := range []struct {
			name string
			want []byte
		}{{"before", originals[member.Path]}, {"members", member.Content}} {
			raw, err := os.ReadFile(filepath.Join(directory, item.name, filepath.FromSlash(member.Path)))
			if err != nil || !bytes.Equal(raw, item.want) {
				t.Fatalf("export lost bytes: %s/%s: %v", item.name, member.Path, err)
			}
		}
	}
}

func TestProfileCLIInputBoundariesFresh1855(t *testing.T) {
	for _, item := range []struct {
		name string
		args []string
	}{
		{"missing-mode", []string{"--request", "missing", "--source-after", "missing"}},
		{"mixed-mode", []string{"--request", "missing", "--source-after", "missing", "--inspect", "--output", "x"}},
		{"unknown-flag", []string{"--apply"}},
	} {
		t.Run(item.name, func(t *testing.T) {
			if err := runProfileMigration(item.args, io.Discard, io.Discard); err == nil {
				t.Fatal("invalid CLI contract accepted")
			}
		})
	}
	t.Run("source-limit", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "large.gooo")
		profileCLIWrite1854(t, path, make([]byte, (128<<10)+1))
		if _, err := readProfileFile(path, 128<<10); err == nil {
			t.Fatal("unbounded replacement accepted")
		}
	})
}

func TestProfileCLIOutputBoundariesFresh1856(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "input")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"inside", "symlink-parent", "existing"} {
		t.Run(name, func(t *testing.T) {
			output := filepath.Join(root, "proposal")
			if name == "symlink-parent" {
				link := filepath.Join(base, "link")
				if err := os.Symlink(root, link); err != nil {
					t.Fatal(err)
				}
				output = filepath.Join(link, "proposal")
			} else if name == "existing" {
				output = filepath.Join(base, "exists")
				if err := os.Mkdir(output, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := profileOutputDirectory(root, output); err == nil {
				t.Fatal("invalid output destination accepted")
			}
		})
	}
}
