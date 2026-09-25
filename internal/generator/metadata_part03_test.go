package generator

import (
	"bytes"
	"testing"
)

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
	mutatedIR := cloneProjectionV1(first)
	mutatedIR.SemanticIR.Package = "tampered"
	if _, err := mutatedIR.CanonicalJSON(); err == nil {
		t.Fatal("SemanticIR digest mismatch was accepted")
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
