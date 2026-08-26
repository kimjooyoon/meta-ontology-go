package generator

import (
	"encoding/json"
	"fmt"
)

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
