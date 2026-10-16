package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func transportFixtureFresh1677(t *testing.T, names []string) config {
	t.Helper()
	directory := t.TempDir()
	cfg := config{Input: filepath.Join(directory, "unpacked"), Metadata: filepath.Join(directory, "metadata.json"),
		Archive: filepath.Join(directory, "producer.zip"), Origin: origin{Repository: "example/gooo",
			SubjectSHA: strings.Repeat("a", 40), RunID: 77, Attempt: 1, ArtifactID: 88}}
	if err := os.Mkdir(cfg.Input, 0750); err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, name := range names {
		member, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		content := []byte("transport:" + name)
		if _, err := member.Write(content); err != nil {
			t.Fatal(err)
		}
		if filepath.Base(name) == name {
			writeFixtureFresh1677(t, filepath.Join(cfg.Input, name), content)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	cfg.Origin.ArtifactDigest = artifact.ContentDigest(buffer.Bytes())
	writeFixtureFresh1677(t, cfg.Archive, buffer.Bytes())
	metadata := artifactMetadata{ID: 88, Name: "metric-concept-operation-" + cfg.Origin.SubjectSHA + "-77-1",
		Digest: cfg.Origin.ArtifactDigest}
	metadata.WorkflowRun.ID, metadata.WorkflowRun.HeadSHA = 77, cfg.Origin.SubjectSHA
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFresh1677(t, cfg.Metadata, raw)
	return cfg
}

func writeFixtureFresh1677(t *testing.T, path string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(path, raw, 0640); err != nil {
		t.Fatal(err)
	}
}
