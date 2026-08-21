package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func validateProvenanceEvidence(records []provenance.Evidence, sourceDigest, semanticDigest, graphDigest string) error {
	if len(records) == 0 {
		return errors.New("evidence is incomplete: at least one record is required")
	}
	for index := range records {
		record := &records[index]
		if strings.ToLower(strings.TrimSpace(record.SourceDigest)) != sourceDigest {
			return fmt.Errorf("evidence record %d has source digest different from authoritative input", index)
		}
		if strings.ToLower(strings.TrimSpace(record.SemanticDigest)) != semanticDigest {
			return fmt.Errorf("evidence record %d has semantic digest different from authoritative source", index)
		}
		if strings.ToLower(strings.TrimSpace(record.GraphDigest)) != graphDigest {
			return fmt.Errorf("evidence record %d has graph digest different from authoritative source", index)
		}
		parsed, err := semantic.ParseIdentity(strings.TrimSpace(record.SemanticID))
		if err != nil {
			return fmt.Errorf("evidence record %d has invalid stable semantic ID: %w", index, err)
		}
		record.SemanticID = parsed.String()
	}
	return nil
}
func verifiedClaims(records []provenance.Evidence) []provenance.VerifiedClaim {
	claims := make([]provenance.VerifiedClaim, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record.Status != provenance.StatusVerified {
			continue
		}
		key := record.SemanticID + "\x00" + record.SemanticDigest + "\x00" + record.GraphDigest
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		claims = append(claims, provenance.VerifiedClaim{
			SemanticID: record.SemanticID, SemanticDigest: record.SemanticDigest, GraphDigest: record.GraphDigest,
		})
	}
	return claims
}
func validatePublishedSnapshot(snapshot provenance.Snapshot, input []provenance.Evidence, sourceDigest, semanticDigest, graphDigest string) error {
	byID := make(map[string]provenance.Evidence, len(snapshot.Records))
	for _, record := range snapshot.Records {
		if record.SourceDigest != sourceDigest || record.SemanticDigest != semanticDigest || record.GraphDigest != graphDigest {
			return errors.New("canonical provenance reread is not bound to the authoritative source")
		}
		byID[record.ID] = record
	}
	for _, record := range input {
		if _, ok := byID[record.ID]; !ok {
			return fmt.Errorf("canonical provenance reread omitted event %q", record.ID)
		}
	}
	return nil
}
