package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

type ExperienceMemoryFingerprint = provenance.ExperienceMemoryFingerprint
type ExperienceMemoryObservationStatus = provenance.ExperienceMemoryObservationStatus
type ExperienceMemoryReceipt = provenance.ExperienceMemoryReceipt

const (
	ExperienceMemoryReceiptSchema = provenance.ExperienceMemoryReceiptSchema
	ExperienceMemoryClosed         = provenance.ExperienceMemoryClosed
	ExperienceMemoryUnknown        = provenance.ExperienceMemoryUnknown
	ExperienceMemoryRefuted        = provenance.ExperienceMemoryRefuted
)

func ObserveExperienceMemory(receipt ExperienceMemoryReceipt, expected ExperienceMemoryFingerprint) ExperienceMemoryReceipt {
	return receipt.Observe(expected)
}
