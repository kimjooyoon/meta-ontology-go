package proofchoicejudge

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func lowerSource(path string, source []byte) (lowered, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || len(diagnostics) != 0 {
		return lowered{}, fmt.Errorf("SOURCE_PARSE_UNKNOWN")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return lowered{}, fmt.Errorf("SOURCE_LOWER_UNKNOWN: %w", err)
	}
	return collectValues(ir)
}

type artifact struct {
	Schema         string `json:"schema"`
	SemanticDigest string `json:"semantic_digest"`
	Canonical      string `json:"canonical"`
}

func decodeArtifact(data []byte) (artifact, error) {
	var result artifact
	if err := json.Unmarshal(data, &result); err != nil || result.Schema != "gooo/proof-choice-canonical-artifact/v1" || result.SemanticDigest == "" || result.Canonical == "" {
		return artifact{}, fmt.Errorf("ARTIFACT_UNKNOWN")
	}
	if digestBytes([]byte(result.Canonical)) != result.SemanticDigest {
		return artifact{}, fmt.Errorf("ARTIFACT_DIGEST_MISMATCH")
	}
	return result, nil
}

func makeArtifact(source lowered) artifact {
	return artifact{Schema: "gooo/proof-choice-canonical-artifact/v1", SemanticDigest: source.SemanticDigest, Canonical: source.Canonical}
}
