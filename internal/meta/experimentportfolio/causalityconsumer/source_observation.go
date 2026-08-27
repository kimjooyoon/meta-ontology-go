package causalityconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func ObserveSource(sourcePath string, source []byte) (SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return SourceObservation{}, sourceObserverFailure("parse-gooo", "SOURCE_PARSE_DIAGNOSTICS", nil)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceObservation{}, sourceObserverFailure("lower-gooo", "SOURCE_SEMANTIC_LOWERING_UNAVAILABLE", err)
	}
	if err := ir.Validate(); err != nil {
		return SourceObservation{}, sourceObserverFailure("validate-semantic-ir", "SOURCE_SEMANTIC_PROJECTION_INVALID", err)
	}
	values := make([]string, 0, 1)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.ValueProgram != "" {
			values = append(values, node.ValueProgram)
		}
	}
	if len(values) == 0 {
		return SourceObservation{}, sourceObserverFailure("project-semantic-value", "SOURCE_SEMANTIC_VALUE_UNAVAILABLE", nil)
	}
	if len(values) != 1 {
		return SourceObservation{}, sourceObserverFailure("project-semantic-value", "SOURCE_SEMANTIC_VALUE_AMBIGUOUS", nil)
	}
	return SourceObservation{
		SourcePath:    canonicalSourcePath(sourcePath),
		SourceDigest:  sha256Digest(source),
		SemanticValue: values[0],
	}, nil
}

func sourceObserverFailure(step, reason string, cause error) error {
	if cause == nil {
		return fmt.Errorf("FAIL_CLOSED: stage=SOURCE_OBSERVER step=%s reason=%s", step, reason)
	}
	return fmt.Errorf("FAIL_CLOSED: stage=SOURCE_OBSERVER step=%s reason=%s: %w", step, reason, cause)
}

func canonicalSourcePath(path string) string {
	path = filepath.ToSlash(path)
	const marker = "examples/experiment-portfolio/"
	if index := strings.Index(path, marker); index >= 0 {
		return path[index:]
	}
	return path
}

func sha256Digest(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}
