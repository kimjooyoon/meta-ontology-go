package generator

import (
	"fmt"
)

// GenerateWithBinding verifies an opt-in authoritative binding without
// persisting it. Missing external references remain explicitly deferred.
func GenerateWithBinding(ir SemanticIR, previous []byte, binding ProjectionBinding) (ProjectionMetadataV1, error) {
	result, err := GenerateProjectionV1(ir, previous)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	if err := validateProjectionBinding(result, ir, binding); err != nil {
		return ProjectionMetadataV1{}, err
	}
	result.Metadata.Source = BindingStatus{Status: "BOUND", Authority: "caller-binding"}
	result.Metadata.SemanticIR = BindingStatus{Status: "BOUND", Authority: "caller-binding"}
	result.Metadata.Projection = ProjectionStatus{Decision: "PASS", Refs: []string{"go-generator", "caller-binding"}}
	if binding.EvidenceDigest != "" {
		result.Metadata.Evidence = EvidenceStatus{Decision: "UNVERIFIED", Refs: []string{binding.EvidenceDigest}}
	}
	if binding.ProvenanceDigest != "" {
		result.Metadata.Provenance = BindingStatus{Status: "UNVERIFIED", Authority: "caller-supplied-unverified"}
	}
	if binding.Toolchain.Value != "" {
		result.Metadata.Toolchain = ToolchainIdentity{Status: "UNVERIFIED", Value: binding.Toolchain.Value}
	}
	result.Metadata.Authority.Provenance = "caller-supplied-unverified"
	return result, nil
}
func validateProjectionBinding(result ProjectionMetadataV1, ir SemanticIR, binding ProjectionBinding) error {
	if binding.Schema != projectionBindingSchemaV1 {
		return fmt.Errorf("generator: unsupported projection binding schema %q", binding.Schema)
	}
	expectedIR := digestIR(result.SemanticIR)
	if expectedIR == "" || expectedIR != digestIR(ir) {
		return fmt.Errorf("generator: projection SemanticIR digest is not normalized")
	}
	if !validDigest(binding.SourceDigest) || binding.SourceDigest != digestBytes(result.Source) {
		return fmt.Errorf("generator: projection binding source digest mismatch")
	}
	if !validDigest(binding.SemanticIRDigest) || binding.SemanticIRDigest != expectedIR {
		return fmt.Errorf("generator: projection binding SemanticIR digest mismatch")
	}
	if !validDigest(binding.SourceMapDigest) || binding.SourceMapDigest != digestSourceMap(result.SourceMap) {
		return fmt.Errorf("generator: projection binding SourceMap digest mismatch")
	}
	if binding.EvidenceDigest != "" && !validDigest(binding.EvidenceDigest) {
		return fmt.Errorf("generator: invalid evidence digest")
	}
	if binding.ProvenanceDigest != "" && !validDigest(binding.ProvenanceDigest) {
		return fmt.Errorf("generator: invalid provenance digest")
	}
	if binding.Toolchain.Value != "" && binding.Toolchain.Status != "BOUND" {
		return fmt.Errorf("generator: supplied toolchain identity must be BOUND")
	}
	return nil
}
func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	for _, char := range value[len("sha256:"):] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
