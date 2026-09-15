package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func semanticCheckEvidence(filename string, file *syntax.File, sourceDigest, semanticDigest, graphDigest string) provenance.Evidence {
	identity := semanticCheckEvidenceIdentity(sourceDigest, semanticDigest, graphDigest)
	span := syntaxFileSpan(file)
	if span.Filename == "" {
		span.Filename = filename
	}
	return provenance.Evidence{
		ID: identity, SemanticID: identity, Producer: string(semantic.GoHostedCompilerID),
		Kind: provenance.KindCompilerRun, Status: provenance.StatusVerified,
		SourceSpan: provenance.SourceSpan{
			URI:   span.Filename,
			Start: provenance.Position{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
			End:   provenance.Position{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
		},
		SourceDigest: sourceDigest, SemanticDigest: semanticDigest, GraphDigest: graphDigest,
		Attributes: map[string]string{
			"check_schema": semanticCheckSchemaVersion,
			"result":       "pass",
		},
		Freshness: provenance.NewFreshness(sourceDigest, time.Unix(0, 0).UTC(), time.Time{}),
	}
}
func semanticCheckEvidenceIdentity(sourceDigest, semanticDigest, graphDigest string) string {

	canonical := strings.Join([]string{
		"provenance/v1", semanticCheckSchemaVersion,
		string(provenance.KindCompilerRun), string(provenance.StatusVerified),
		sourceDigest, semanticDigest, graphDigest,
	}, "\x00")
	return "gooo://event/" + semantic.StableHashString(canonical)
}
func validateProvenanceStoreParent(storePath string) error {
	parent := filepath.Dir(filepath.Clean(storePath))
	info, err := os.Stat(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("provenance store parent does not exist: %s", parent)
		}
		return fmt.Errorf("inspect provenance store parent: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("provenance store parent is not a directory: %s", parent)
	}
	return nil
}
