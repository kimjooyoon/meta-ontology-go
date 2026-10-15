package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"testing"
)

type namedInputArchiveManifest struct {
	Archive string `json:"archive"`
	SHA256  string `json:"archive_sha256"`
	Members []struct {
		Path   string `json:"path"`
		Bytes  int    `json:"bytes"`
		SHA256 string `json:"sha256"`
	} `json:"members"`
}

func TestNamedInputSavedSourceBindingFirst(t *testing.T) {
	const directory = "../../docs/observations/named-inputs/"
	raw, err := os.ReadFile(directory + "archive-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest namedInputArchiveManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(directory + manifest.Archive)
	if err != nil {
		t.Fatal(err)
	}
	if namedInputObservationSHA(archive) != manifest.SHA256 {
		t.Fatal("archive differs")
	}
	entries := readNamedInputArchive(t, archive)
	if len(entries) != len(manifest.Members) {
		t.Fatal("archive membership differs")
	}
	for _, member := range manifest.Members {
		raw, ok := entries[member.Path]
		if !ok || len(raw) != member.Bytes || namedInputObservationSHA(raw) != member.SHA256 {
			t.Fatal("saved source/observation differs", member.Path)
		}
	}
	const source = "examples/named-inputs/main.gooo"
	current, err := os.ReadFile("../../" + source)
	if err != nil || !bytes.Equal(current, entries["source/"+source]) {
		t.Fatal("example source differs", err)
	}
}

func readNamedInputArchive(t *testing.T, raw []byte) map[string][]byte {
	t.Helper()
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	reader := tar.NewReader(z)
	entries := map[string][]byte{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if header.Typeflag != tar.TypeReg {
			t.Fatal("unexpected archive member", header.Name)
		}
		if _, exists := entries[header.Name]; exists {
			t.Fatal("duplicate member", header.Name)
		}
		data, err := io.ReadAll(reader)
		if err != nil || int64(len(data)) != header.Size {
			t.Fatal("incomplete member", header.Name, err)
		}
		entries[header.Name] = data
	}
	return entries
}

func namedInputObservationSHA(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
