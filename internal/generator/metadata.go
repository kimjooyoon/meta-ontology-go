package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const projectionMetadataSchemaV1 = "gooo-generator/v1"
const projectionBindingSchemaV1 = "gooo-generator-binding/v1"

// GenerateWithMetadata is a companion API; Generate and GenerateFrom remain
// unchanged for existing callers.
func GenerateWithMetadata(ir SemanticIR, previous []byte) (MetadataResult, error) {
	result, err := Generate(ir, previous)
	if err != nil {
		return MetadataResult{}, err
	}
	return metadataResult(result, ir), nil
}

// GenerateProjectionV1 returns a versioned, canonicalizable result surface.
func GenerateProjectionV1(ir SemanticIR, previous []byte) (ProjectionMetadataV1, error) {
	return generateProjectionV1(New(Options{}), ir, previous)
}

func generateProjectionV1(generator Generator, ir SemanticIR, previous []byte) (ProjectionMetadataV1, error) {
	return generateProjectionV1WithEntityFieldsSupport(generator, ir, previous, checkedEntityFieldsSupport())
}

// generateProjectionV1WithEntityFieldsSupport is package-private so focused
// tests can prove the exact profile-bound SUPPORTED branch without exposing a
// caller-selectable production activation surface.
func generateProjectionV1WithEntityFieldsSupport(generator Generator, ir SemanticIR, previous []byte, support entityFieldsSupport) (ProjectionMetadataV1, error) {
	if err := validateEntityFieldsInput(ir, support); err != nil {
		return ProjectionMetadataV1{}, err
	}
	working := ir
	if support.State == entityFieldsSupported && semanticIRHasFields(ir) {
		working = prepareEntityFields(ir)
	}
	normalized, err := normalizeIR(working)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	result, err := generator.generateWithEntityFieldsSupport(normalized, previous, support)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	metadata := metadataResultWithEntityFieldsSupport(result, normalized, support)
	return ProjectionMetadataV1{
		Schema:     projectionMetadataSchemaV1,
		Source:     append([]byte(nil), result.Source...),
		SemanticIR: normalized,
		SourceMap:  result.SourceMap,
		Metadata:   metadata.Metadata,
	}, nil
}

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

// CanonicalJSON returns deterministic JSON and verifies its embedded digests.
func (result ProjectionMetadataV1) CanonicalJSON() ([]byte, error) {
	if result.Schema != projectionMetadataSchemaV1 {
		return nil, fmt.Errorf("generator: unsupported projection metadata schema %q", result.Schema)
	}
	if !validDigest(result.Metadata.SourceDigest) || result.Metadata.SourceDigest != digestBytes(result.Source) ||
		!validDigest(result.Metadata.SemanticIRDigest) || result.Metadata.SemanticIRDigest != digestIR(result.SemanticIR) ||
		!validDigest(result.Metadata.SourceMapDigest) || result.Metadata.SourceMapDigest != digestSourceMap(result.SourceMap) {
		return nil, fmt.Errorf("generator: projection metadata digest mismatch")
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("generator: marshal projection metadata: %w", err)
	}
	return payload, nil
}

// CanonicalHash returns the digest of CanonicalJSON without writing a receipt.
func (result ProjectionMetadataV1) CanonicalHash() (string, error) {
	payload, err := result.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}

func metadataResult(result Result, ir SemanticIR) MetadataResult {
	return MetadataResult{
		Result: result,
		Metadata: GenerationMetadata{
			SourceDigest:     digestBytes(result.Source),
			SemanticIRDigest: digestIR(ir),
			SourceMapDigest:  digestSourceMap(result.SourceMap),
			Source:           BindingStatus{Status: "AVAILABLE", Authority: "generator-output"},
			SemanticIR:       BindingStatus{Status: "AVAILABLE", Authority: "generator-input"},
			Provenance:       BindingStatus{Status: "DEFERRED", Authority: "external-receipt-required"},
			Evidence: EvidenceStatus{
				Decision: "DEFERRED",
				Refs:     []string{},
			},
			Toolchain: ToolchainIdentity{
				Status: "DEFERRED",
				Value:  "",
			},
			Projection: ProjectionStatus{
				Decision: "PASS",
				Refs:     []string{"go-generator"},
			},
			Authority: AuthorityLabels{
				Projection: "go-generator",
				Verifier:   "go-verifier-stage-0",
				Provenance: "external-receipt-required",
			},
		},
	}
}

func metadataResultWithEntityFieldsSupport(result Result, ir SemanticIR, support entityFieldsSupport) MetadataResult {
	metadata := metadataResult(result, ir)
	if semanticIRHasFields(ir) {
		metadata.Metadata.EntityFields = entityFieldsMetadata(support)
	}
	return metadata
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestIR(ir SemanticIR) string {
	normalized, err := normalizeIR(ir)
	if err != nil {
		return ""
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return ""
	}
	return digestBytes(payload)
}

func digestSourceMap(sourceMap SourceMap) string {
	payload, err := json.Marshal(sourceMap)
	if err != nil {
		return ""
	}
	return digestBytes(payload)
}
