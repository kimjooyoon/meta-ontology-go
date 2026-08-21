package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func provenanceCLIRecords(records []provenance.Evidence) []provenanceCLIRecord {
	result := make([]provenanceCLIRecord, 0, len(records))
	for _, record := range records {
		result = append(result, provenanceCLIRecord{
			ID: record.ID, SemanticID: record.SemanticID, Producer: record.Producer,
			Kind: string(record.Kind), Status: string(record.Status), Sequence: record.Sequence, Hash: record.Hash,
		})
	}
	return result
}
