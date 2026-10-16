package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestProfileCLIPinnedRequestFresh1860(t *testing.T) {
	root, requestPath, afterPath, directory, _ := profileCLIFixture1854(t)
	args := []string{"--root", root, "--request", requestPath, "--source-after", afterPath}
	var output bytes.Buffer
	if err := runProfileMigration(append(args, "--inspect"), &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	var response struct {
		Request syntaxregistration.ProfileMigrationRequest `json:"request"`
	}
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Request.Case.ID != "cli1854" || response.Request.SnapshotDigest == "" ||
		response.Request.SourceDigest == "" || response.Request.ReplacementDigest == "" ||
		response.Request.ExecutionIdentity.ExecutableDigest == "" || response.Request.Toolchain == "" {
		t.Fatal("host must reconstruct or infer a typed request")
	}
	raw, err := json.Marshal(response.Request)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(requestPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := runProfileMigration(append(args, "--output", directory), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
}
