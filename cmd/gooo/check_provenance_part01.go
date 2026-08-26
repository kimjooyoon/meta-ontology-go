package main

import (
	"errors"
	"fmt"
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
