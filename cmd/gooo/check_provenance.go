package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const checkStatusPass = "pass"

// publishSemanticCheckProvenance records only a successful, fully validated
// semantic check. The evidence event describes itself as its semantic subject:
// the check does not invent a business node or promote a candidate fact.
func publishSemanticCheckProvenance(filename string, source []byte, file *syntax.File, ir semantic.IR, storePath string) (provenancePublishResponse, error) {
	response := provenancePublishResponse{
		Schema: provenanceCLISchema, Status: provenanceStatusRejected, CheckStatus: checkStatusPass, Records: []provenanceCLIRecord{},
		SourceDigest: semantic.StableHash(source), SemanticDigest: ir.StableHash(), GraphDigest: ir.Graph.StableHash(),
	}
	if len(ir.Graph.Candidates()) != 0 {
		return response, errors.New("semantic check evidence is incomplete: candidate facts are not verified")
	}
	if err := validateProvenanceStoreParent(storePath); err != nil {
		return response, err
	}
	record := semanticCheckEvidence(filename, file, response.SourceDigest, response.SemanticDigest, response.GraphDigest)
	store := provenance.New(storePath)
	existing, err := store.Read(provenance.ReadOptions{ExpectedSourceDigest: response.SourceDigest})
	if err != nil {
		return response, fmt.Errorf("canonical provenance preflight: %w", err)
	}
	for _, prior := range existing.Records {
		if prior.SemanticDigest != response.SemanticDigest || prior.GraphDigest != response.GraphDigest {
			return response, errors.New("canonical provenance preflight is not bound to the authoritative semantic graph")
		}
	}
	if err := store.Append(record); err != nil {
		return response, fmt.Errorf("append provenance: %w", err)
	}
	snapshot, err := store.Read(provenance.ReadOptions{
		ExpectedSourceDigest: response.SourceDigest,
		RequiredVerified: []provenance.VerifiedClaim{{
			SemanticID: record.SemanticID, SemanticDigest: response.SemanticDigest, GraphDigest: response.GraphDigest,
		}},
	})
	if err != nil {
		return response, fmt.Errorf("canonical provenance reread: %w", err)
	}
	if err := validatePublishedSnapshot(snapshot, []provenance.Evidence{record}, response.SourceDigest, response.SemanticDigest, response.GraphDigest); err != nil {
		return response, err
	}
	response.StoreDigest = snapshot.Digest
	response.Records = provenanceCLIRecords(snapshot.Records)
	response.Status = provenanceStatusCommitted
	if err := sealProvenanceResponse(&response); err != nil {
		return response, fmt.Errorf("seal provenance response: %w", err)
	}
	return response, nil
}

func rejectSemanticCheckProvenance(response provenancePublishResponse, cause error) (provenancePublishResponse, error) {
	response.Status = provenanceStatusRejected
	response.StoreDigest = ""
	response.Records = []provenanceCLIRecord{}
	response.Error = &provenanceCLIError{Code: provenanceErrorCode(cause), Message: cause.Error()}
	if err := sealProvenanceResponse(&response); err != nil {
		return response, fmt.Errorf("seal provenance response: %w", err)
	}
	return response, nil
}

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
	// Event identity is content-bound to the contract and exact source/IR/graph
	// snapshots; producer, source location, and freshness remain audit fields.
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
