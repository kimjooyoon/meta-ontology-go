package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func evidenceForFact(
	records []semantic.Evidence, key semantic.FactKey, status semantic.FactStatus,
) (semantic.Evidence, bool) {
	for _, record := range records {
		if record.Fact == key && record.Status == status {
			return record, true
		}
	}
	return semantic.Evidence{}, false
}
func shadowEvidenceForFact(records []semantic.Evidence, key semantic.FactKey) (semantic.Evidence, bool) {
	return evidenceForFact(records, key, semantic.FactCandidate)
}
