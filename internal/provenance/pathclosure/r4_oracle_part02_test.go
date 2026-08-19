package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// plainFiniteGraphBaseline is deliberately weaker than R4: it checks only a
// finite ordered graph and never pretends to validate receipts or phases.
func plainFiniteGraphBaseline(input pathclosure.R4Input) bool {
	if input.Boundary.OpenWorld || !input.Boundary.Exhausted || len(input.Boundary.RequiredPathIDs) == 0 {
		return false
	}
	records := map[semantic.ID]pathclosure.R4Record{}
	for _, record := range input.Records {
		records[record.ID] = record
	}
	paths := map[semantic.ID]pathclosure.R4Path{}
	for _, path := range input.Paths {
		paths[path.ID] = path
	}
	for _, required := range input.Boundary.RequiredPathIDs {
		path, ok := paths[required]
		if !ok || len(path.RecordIDs) == 0 || len(path.RecordIDs) != len(path.RecordBytes) {
			return false
		}
		var previous pathclosure.R4Record
		for index, recordID := range path.RecordIDs {
			record, ok := records[recordID]
			if !ok {
				return false
			}
			if index == 0 {
				if record.PredecessorID != "" || record.SubjectID != path.StartID {
					return false
				}
			} else if record.PredecessorID != previous.ID || previous.ObjectID != record.SubjectID {
				return false
			}
			if index == len(path.RecordIDs)-1 && record.ObjectID != path.EndID {
				return false
			}
			previous = record
		}
	}
	return true
}
