package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const ArtifactSchema = "gooo/capability-scoped-expansion/artifact/v1"

type Request struct {
	SourceSemanticDigest string
	GraphPathDigest      string
	FileDigest           string
	LogicalValue         string
	OutputPath           string
}

type Artifact struct {
	Schema         string
	Path           string
	Value          string
	Bytes          []byte
	ContentDigest  string
	SemanticDigest string
}

// Expand is a read-only expansion engine. It has no effects dependency; the
// only effect API in this experiment requires a broker.Token and is therefore
// unreachable from this package's import closure.
func Expand(request Request) (Artifact, error) {
	if request.SourceSemanticDigest == "" || request.GraphPathDigest == "" || request.FileDigest == "" || request.LogicalValue == "" || request.OutputPath == "" {
		return Artifact{}, fmt.Errorf("expansion input is incomplete")
	}
	value := fmt.Sprintf("capability.expanded|source=%s|graph=%s|file-digest=%s|logical=%s", request.SourceSemanticDigest, request.GraphPathDigest, request.FileDigest, request.LogicalValue)
	source := fmt.Sprintf("package capabilityscopedexpansion\nnamespace capabilityscopedexpansion\n\nentity ExpandedSyntax id \"gooo://capability-scoped-expansion/expanded\"\nactivity EmitExpandedSyntax(ExpandedSyntax) -> ExpandedSyntax computes %q\n", value)
	bytes := []byte(source)
	file, diagnostics := syntax.ParseFile("expanded-capability.gooo", source)
	if err := diagnostics.Error(); err != nil {
		return Artifact{}, err
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Artifact{}, err
	}
	if err := os.MkdirAll(filepath.Dir(request.OutputPath), 0o755); err != nil {
		return Artifact{}, err
	}
	if err := os.WriteFile(request.OutputPath, bytes, 0o644); err != nil {
		return Artifact{}, err
	}
	written, err := os.ReadFile(request.OutputPath)
	if err != nil {
		return Artifact{}, err
	}
	if string(written) != source {
		return Artifact{}, fmt.Errorf("expanded artifact readback changed")
	}
	return Artifact{Schema: ArtifactSchema, Path: request.OutputPath, Value: value, Bytes: bytes, ContentDigest: digest(bytes), SemanticDigest: ir.StableHash()}, nil
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
