package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateWithMetadataIsDeterministic(t *testing.T) {
	ir := acceptanceFixture()
	first, err := GenerateWithMetadata(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithMetadata(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Source, second.Source) || !reflect.DeepEqual(first.Metadata, second.Metadata) {
		t.Fatal("metadata result is not deterministic")
	}
	for _, digest := range []string{first.Metadata.SourceDigest, first.Metadata.SemanticIRDigest, first.Metadata.SourceMapDigest} {
		if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
			t.Fatalf("invalid digest %q", digest)
		}
	}
}

func TestGenerateWithMetadataMarksUnavailableEvidenceAndToolchain(t *testing.T) {
	result, err := GenerateWithMetadata(acceptanceFixture(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.Evidence.Decision != "DEFERRED" || len(result.Metadata.Evidence.Refs) != 0 {
		t.Fatalf("unexpected evidence status: %#v", result.Metadata.Evidence)
	}
	if result.Metadata.Toolchain.Status != "DEFERRED" || result.Metadata.Toolchain.Value != "" {
		t.Fatalf("unexpected toolchain identity: %#v", result.Metadata.Toolchain)
	}
	if result.Metadata.Projection.Decision != "PASS" {
		t.Fatalf("unexpected projection status: %#v", result.Metadata.Projection)
	}
}

func TestGenerateWithMetadataDoesNotMutateIROrPrevious(t *testing.T) {
	ir := acceptanceFixture()
	beforeIR := copyIR(ir)
	initial, err := Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	previous := append([]byte(nil), initial.Source...)
	beforePrevious := append([]byte(nil), previous...)
	if _, err := GenerateWithMetadata(ir, previous); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ir, beforeIR) {
		t.Fatal("metadata generation mutated caller-owned IR")
	}
	if !bytes.Equal(previous, beforePrevious) {
		t.Fatal("metadata generation mutated caller-owned previous source")
	}
}
