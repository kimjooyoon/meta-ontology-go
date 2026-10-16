package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

func TestOriginArchiveCLIRefusalFresh1817(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runViewLens([]string{"read", "--state-dir", t.TempDir(), "--kind", "profile", "--digest", "invalid"},
		&stdout, &stderr)
	var got viewlens.ArchiveRead
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil || code != exitOK ||
		got.State != "UNAVAILABLE" || got.Issue != "ARCHIVE_REQUEST_INVALID" || len(got.Raw) != 0 {
		t.Fatalf("detached archive diagnostic: %v %d %s %s", err, code, &stdout, &stderr)
	}
}
