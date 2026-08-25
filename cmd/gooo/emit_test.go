package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func TestRunEmitProjectsOperationManifest(t *testing.T) {
	var stdout, stderr bytes.Buffer
	directory := filepath.Join("..", "..", "examples", "billing-package")
	code := runEmit([]string{"--kind", "operation-manifest", "--entry", "PayOrder", directory}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var artifact artifactemit.Artifact
	if err := json.Unmarshal(stdout.Bytes(), &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.Decision != "PASS" || artifact.Operation.Activity != "PayOrder" ||
		len(artifact.Definitions.Files) != 2 || artifact.Extensions.RegisteredEmitters != 1 {
		t.Fatalf("artifact=%#v", artifact)
	}
}

func TestRunEmitRejectsUnknownEmitter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	directory := filepath.Join("..", "..", "examples", "billing-package")
	code := runEmit([]string{"--kind", "not-registered", "--entry", "PayOrder", directory}, &stdout, &stderr)
	var artifact artifactemit.Artifact
	if err := json.Unmarshal(stdout.Bytes(), &artifact); err != nil {
		t.Fatal(err)
	}
	if code != exitFailure || stderr.Len() != 0 || artifact.Decision != "FAIL_CLOSED" ||
		artifact.Resolution != "LOWER_RESOLUTION" || artifact.Reason != "EMITTER_UNKNOWN" {
		t.Fatalf("code=%d stderr=%q artifact=%#v", code, stderr.String(), artifact)
	}
}
