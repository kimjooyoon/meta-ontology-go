package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestCandidateExportRefusesInputAndExistingDirectories(t *testing.T) {
	root := t.TempDir()
	for _, output := range []string{root, filepath.Join(root, "candidate")} {
		if err := exportCandidate(root, output, syntaxregistration.Candidate{}, 0); err == nil {
			t.Fatalf("input repository output was accepted: %s", output)
		}
	}
	existing := t.TempDir()
	if err := exportCandidate(root, existing, syntaxregistration.Candidate{}, 0); err == nil {
		t.Fatal("existing caller-owned output was overwritten")
	}
}

func TestCandidateExportResolvesSymlinkParents(t *testing.T) {
	root := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "input-link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := exportCandidate(root, filepath.Join(link, "candidate"), syntaxregistration.Candidate{}, 0); err == nil {
		t.Fatal("symlink parent escaped the no-input-writes boundary")
	}
}
