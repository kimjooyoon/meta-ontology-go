package generator

import (
	"bytes"
	"reflect"
	"testing"
)

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
