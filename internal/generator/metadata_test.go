package generator

import (
	"bytes"
	"encoding/json"
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
	if result.Metadata.Source.Status != "AVAILABLE" || result.Metadata.SemanticIR.Status != "AVAILABLE" {
		t.Fatalf("source/IR bindings are not available: %#v", result.Metadata)
	}
	if result.Metadata.Provenance.Status != "DEFERRED" || result.Metadata.Authority.Verifier != "go-verifier-stage-0" {
		t.Fatalf("authority boundary was not explicit: %#v", result.Metadata)
	}
}

func TestGenerateWithMetadataJSONIsCanonicalAcrossPermutation(t *testing.T) {
	firstIR := acceptanceFixture()
	secondIR := acceptanceFixture()
	secondIR.Entities[0], secondIR.Entities[1] = secondIR.Entities[1], secondIR.Entities[0]
	first, err := GenerateWithMetadata(firstIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithMetadata(secondIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := json.Marshal(first.Metadata)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second.Metadata)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatal("metadata JSON changed under declaration permutation")
	}
}

func TestGenerateWithMetadataRejectsInvalidIRWithoutMetadata(t *testing.T) {
	ir := acceptanceFixture()
	ir.Activities[0].Slots[0].ID = ""
	if result, err := GenerateWithMetadata(ir, nil); err == nil || result.Metadata.SourceDigest != "" {
		t.Fatalf("invalid IR returned metadata: result=%#v err=%v", result, err)
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
