package artifactcoverage

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalProgramIsMetaBound(t *testing.T) {
	program := CanonicalProgram()
	if err := Validate(program); err != nil {
		t.Fatal(err)
	}
	if len(program.ArtifactBindings) != 7 {
		t.Fatalf("artifact bindings = %d, want 7", len(program.ArtifactBindings))
	}
}

func TestCanonicalAuthorityCoversActivities(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	digest, err := AuthorityDigest(root, CanonicalProgram())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Fatalf("authority digest = %q", digest)
	}
}

func TestDuplicateOperationBindingFailsClosed(t *testing.T) {
	program := CanonicalProgram()
	program.ArtifactBindings = append(program.ArtifactBindings, program.ArtifactBindings[0])
	if err := Validate(program); err == nil {
		t.Fatal("duplicate operation binding was accepted")
	}
}
