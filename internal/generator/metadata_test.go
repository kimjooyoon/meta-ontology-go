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

func TestProjectionMetadataV1CanonicalHashAndDigestSensitivity(t *testing.T) {
	first, err := GenerateProjectionV1(acceptanceFixture(), nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateProjectionV1(acceptanceFixture(), nil)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := first.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := second.CanonicalJSON()
	if err != nil || !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("canonical JSON is not repeat-stable: %v", err)
	}
	firstHash, err := first.CanonicalHash()
	if err != nil {
		t.Fatal(err)
	}
	mutatedSource := cloneProjectionV1(second)
	mutatedSource.Source[0] ^= 1
	if _, err := mutatedSource.CanonicalJSON(); err == nil {
		t.Fatal("source digest mismatch was accepted")
	}
	mutatedMap := cloneProjectionV1(first)
	mutatedMap.SourceMap.Mappings[0].Ordinal++
	if _, err := mutatedMap.CanonicalJSON(); err == nil {
		t.Fatal("source-map digest mismatch was accepted")
	}
	unchangedHash, err := first.CanonicalHash()
	if err != nil || unchangedHash != firstHash {
		t.Fatalf("canonical hash changed without input mutation: %q %v", unchangedHash, err)
	}
}

func cloneProjectionV1(input ProjectionMetadataV1) ProjectionMetadataV1 {
	output := input
	output.Source = append([]byte(nil), input.Source...)
	output.SourceMap.Mappings = append([]SourceMapping(nil), input.SourceMap.Mappings...)
	output.Metadata.Evidence.Refs = append([]string(nil), input.Metadata.Evidence.Refs...)
	output.Metadata.Projection.Refs = append([]string(nil), input.Metadata.Projection.Refs...)
	return output
}

func TestProjectionMetadataV1UsesDeferredExternalBindings(t *testing.T) {
	result, err := GenerateProjectionV1(acceptanceFixture(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Schema != projectionMetadataSchemaV1 || result.Metadata.Provenance.Status != "DEFERRED" {
		t.Fatalf("unexpected versioned metadata: %#v", result)
	}
	if result.Metadata.Evidence.Decision != "DEFERRED" || result.Metadata.Toolchain.Value != "" {
		t.Fatalf("missing binding was fabricated: %#v", result.Metadata)
	}
}

func TestGenerateWithBindingReplaysAndBinds(t *testing.T) {
	ir := acceptanceFixture()
	base, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	binding := ProjectionBinding{
		Schema: projectionBindingSchemaV1, SourceDigest: base.Metadata.SourceDigest,
		SemanticIRDigest: base.Metadata.SemanticIRDigest, SourceMapDigest: base.Metadata.SourceMapDigest,
		EvidenceDigest: digestBytes([]byte("evidence")), ProvenanceDigest: digestBytes([]byte("provenance")),
		Toolchain: ToolchainIdentity{Status: "BOUND", Value: "go1.26.5"},
	}
	bound, err := GenerateWithBinding(ir, nil, binding)
	if err != nil {
		t.Fatal(err)
	}
	if bound.Metadata.Source.Status != "BOUND" || bound.Metadata.Provenance.Status != "UNVERIFIED" || bound.Metadata.Toolchain.Status != "UNVERIFIED" || bound.Metadata.Toolchain.Value != "go1.26.5" {
		t.Fatalf("binding status not reflected: %#v", bound.Metadata)
	}
	if bound.Metadata.Evidence.Decision != "UNVERIFIED" || bound.Metadata.Authority.Provenance != "caller-supplied-unverified" {
		t.Fatalf("external authority was overstated: %#v", bound.Metadata)
	}
}

func TestGenerateWithBindingRejectsTamperingWithoutMutation(t *testing.T) {
	ir := acceptanceFixture()
	previous, err := Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	binding := ProjectionBinding{Schema: projectionBindingSchemaV1, SourceDigest: digestBytes(previous.Source), SemanticIRDigest: digestIR(ir), SourceMapDigest: digestSourceMap(previous.SourceMap)}
	beforeIR := copyIR(ir)
	beforeSource := append([]byte(nil), previous.Source...)
	binding.SourceDigest = digestBytes([]byte("tampered"))
	if _, err := GenerateWithBinding(ir, previous.Source, binding); err == nil {
		t.Fatal("tampered binding was accepted")
	}
	if !reflect.DeepEqual(ir, beforeIR) || !bytes.Equal(previous.Source, beforeSource) {
		t.Fatal("binding rejection mutated caller-owned inputs")
	}
}

func TestGenerateWithBindingCanonicalJSONCannotPromoteCallerEvidence(t *testing.T) {
	ir := acceptanceFixture()
	base, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := GenerateWithBinding(ir, nil, ProjectionBinding{
		Schema: projectionBindingSchemaV1, SourceDigest: base.Metadata.SourceDigest,
		SemanticIRDigest: base.Metadata.SemanticIRDigest, SourceMapDigest: base.Metadata.SourceMapDigest,
		EvidenceDigest: digestBytes([]byte("forged-receipt")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if bound.Metadata.Evidence.Decision == "BOUND" || bound.Metadata.Provenance.Status == "BOUND" {
		t.Fatal("caller digest was promoted to authoritative evidence")
	}
	if _, err := bound.CanonicalJSON(); err != nil {
		t.Fatal(err)
	}
}
