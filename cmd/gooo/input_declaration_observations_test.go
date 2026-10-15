package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type inputDeclarationObservation struct {
	Source  string            `json:"consumer_source_sha256"`
	Files   map[string]string `json:"observed_files"`
	Archive struct {
		Path    string            `json:"path"`
		SHA256  string            `json:"sha256"`
		Members map[string]string `json:"members"`
	} `json:"archive"`
}

func TestInputDeclarationSavedSourceBinding(t *testing.T) {
	const root = "../.."
	record, saved := readInputDeclarationObservation(t, root)
	const consumer = "examples/source-control-reader/inspect.gooo"
	if fmt.Sprintf("%x", sha256.Sum256(saved["sources/"+consumer])) != record.Source {
		t.Fatal("first consumer source identity lost")
	}
	const fixture = "examples/source-control-reader/input-positions.gooo.fixture"
	source, err := os.ReadFile(filepath.Join(root, fixture))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(source)) != record.Files[fixture] {
		t.Fatal("declared input fixture changed", err)
	}
	var output, diagnostics bytes.Buffer
	code := run([]string{"body-origins", "--activity", "Difference", "--inputs-for", "source/control:InspectInputs",
		"--inputs-shape", "input-declaration-columns", filepath.Join(root, fixture)}, &output, &diagnostics)
	if code != exitOK || !bytes.Equal(output.Bytes(), saved["signature-export-first/stdout"]) {
		t.Fatal("source-only signature export lost its original positions", code, diagnostics.String())
	}
}

func readInputDeclarationObservation(t *testing.T, root string) (inputDeclarationObservation, map[string][]byte) {
	t.Helper()
	const directory = "docs/observations/input-declarations"
	raw, err := os.ReadFile(filepath.Join(root, directory, "observations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record inputDeclarationObservation
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(filepath.Join(root, directory, record.Archive.Path))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(archive)) != record.Archive.SHA256 {
		t.Fatal("observation archive changed", err)
	}
	z, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	return record, readInputDeclarationArchive(t, tar.NewReader(z), record.Archive.Members)
}

func readInputDeclarationArchive(t *testing.T, tr *tar.Reader, members map[string]string) map[string][]byte {
	t.Helper()
	saved := map[string][]byte{}
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(io.LimitReader(tr, 4<<20))
		if err != nil || int64(len(data)) != header.Size || fmt.Sprintf("%x", sha256.Sum256(data)) != members[header.Name] {
			t.Fatal("saved source/input/output changed", header.Name, err)
		}
		if _, exists := saved[header.Name]; exists {
			t.Fatal("duplicate archive member", header.Name)
		}
		saved[header.Name] = data
	}
	if len(saved) != len(members) {
		t.Fatal("missing archive members")
	}
	return saved
}
