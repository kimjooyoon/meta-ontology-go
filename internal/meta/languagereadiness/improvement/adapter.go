package improvement

import readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"

// FromReadiness projects the counted readiness contract into transition input.
func FromReadiness(source readiness.Snapshot) Snapshot {
	result := Snapshot{
		ContractSchema: source.ContractSchema,
		RegistryDigest: source.RegistryDigest,
		Completed:      int64(source.Summary.Completed),
		Total:          int64(source.Summary.Total),
		BasisPoints:    int64(source.Summary.ReadinessBPS),
		Evidence:       make([]Evidence, 0, len(source.Obligations)),
	}
	for _, obligation := range source.Obligations {
		result.Evidence = append(result.Evidence, Evidence{
			ID: obligation.ID, Status: EvidenceStatus(obligation.Status),
		})
	}
	return result
}
