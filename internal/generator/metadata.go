package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const projectionMetadataSchemaV1 = "gooo-generator/v1"

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
	result, err := GenerateWithMetadata(ir, previous)
	if err != nil {
		return ProjectionMetadataV1{}, err
	}
	return ProjectionMetadataV1{
		Schema: projectionMetadataSchemaV1, Source: append([]byte(nil), result.Source...),
		SourceMap: result.SourceMap, Metadata: result.Metadata,
	}, nil
}

// CanonicalJSON returns deterministic JSON and verifies its embedded digests.
func (result ProjectionMetadataV1) CanonicalJSON() ([]byte, error) {
	if result.Schema != projectionMetadataSchemaV1 {
		return nil, fmt.Errorf("generator: unsupported projection metadata schema %q", result.Schema)
	}
	if result.Metadata.SourceDigest != digestBytes(result.Source) || result.Metadata.SourceMapDigest != digestSourceMap(result.SourceMap) {
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
