package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"time"
)

func publishProvenance(options provenancePublishOptions, reader SourceReader, parser SourceParser, deadline time.Time) (provenancePublishResponse, error) {
	response := provenancePublishResponse{
		Schema: provenanceCLISchema, Status: provenanceStatusRejected, Records: []provenanceCLIRecord{},
	}
	source, err := readSourceWithDeadline(reader, options.source, remainingDeadline(deadline))
	if err != nil {
		return response, fmt.Errorf("read source: %w", err)
	}
	response.SourceDigest = semantic.StableHash(source)

	file, diagnostics, err := parseWithDeadline(parser, options.source, string(source), remainingDeadline(deadline))
	if err != nil {
		return response, fmt.Errorf("parse source: %w", err)
	}
	if diagnostics.HasErrors() {
		return response, fmt.Errorf("source diagnostics: %s", diagnostics.Error())
	}
	ir, err := lowerInspectIRWith(file, remainingDeadline(deadline), bidir.Lower)
	if err != nil {
		return response, fmt.Errorf("lower source: %w", err)
	}
	response.SemanticDigest = ir.StableHash()
	response.GraphDigest = ir.Graph.StableHash()

	evidenceBytes, err := readSourceWithDeadline(reader, options.evidence, remainingDeadline(deadline))
	if err != nil {
		return response, fmt.Errorf("read evidence: %w", err)
	}
	records, err := decodeProvenanceEvidence(evidenceBytes)
	if err != nil {
		return response, fmt.Errorf("decode evidence: %w", err)
	}
	if err := validateProvenanceEvidence(records, response.SourceDigest, response.SemanticDigest, response.GraphDigest); err != nil {
		return response, err
	}

	store := provenance.New(options.store)
	if err := store.Append(records...); err != nil {
		return response, fmt.Errorf("append provenance: %w", err)
	}
	claims := verifiedClaims(records)
	snapshot, err := store.Read(provenance.ReadOptions{
		ExpectedSourceDigest: response.SourceDigest,
		RequiredVerified:     claims,
	})
	if err != nil {
		return response, fmt.Errorf("canonical provenance reread: %w", err)
	}
	if err := validatePublishedSnapshot(snapshot, records, response.SourceDigest, response.SemanticDigest, response.GraphDigest); err != nil {
		return response, err
	}
	response.StoreDigest = snapshot.Digest
	response.Records = provenanceCLIRecords(snapshot.Records)
	response.Status = provenanceStatusCommitted
	return response, nil
}
