package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// GenerateWithMetadata is a companion API; Generate and GenerateFrom remain
// unchanged for existing callers.
func GenerateWithMetadata(ir SemanticIR, previous []byte) (MetadataResult, error) {
	result, err := Generate(ir, previous)
	if err != nil {
		return MetadataResult{}, err
	}
	return metadataResult(result, ir), nil
}

func metadataResult(result Result, ir SemanticIR) MetadataResult {
	return MetadataResult{
		Result: result,
		Metadata: GenerationMetadata{
			SourceDigest:     digestBytes(result.Source),
			SemanticIRDigest: digestIR(ir),
			SourceMapDigest:  digestSourceMap(result.SourceMap),
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
