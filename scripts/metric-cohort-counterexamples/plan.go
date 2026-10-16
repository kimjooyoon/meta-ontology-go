package main

type countercase struct {
	ID, Kind, Reason string
	EnvelopeRequired bool
}

const (
	evidenceReasonPrefix = "verify concept-operation evidence: "
	sourceReasonPrefix   = "verify concept-operation source: "
	sourceRebuildReason  = sourceReasonPrefix + "concept-operation receipt does not match independently rebuilt evidence"
	sourceReplayReason   = sourceReasonPrefix + "independent strategy replay: metric strategy independent synthesis diverged"
	missingReceiptReason = evidenceReasonPrefix + "decode concept-operation receipt: EOF"
	identityReason       = evidenceReasonPrefix + "concept-operation receipt identity is invalid"
	programReplayReason  = sourceReasonPrefix + "independent program replay: program does not match independent reconstruction"
	programSourceReason  = sourceReasonPrefix + "independent program replay: Gooo meta program source is not canonical"
)

// The complete case population is declared before any result is observed.
func countercases() []countercase {
	return []countercase{
		{"missing-observed-binding", "PRODUCER_INPUT_MUTATION", sourceReplayReason, false},
		{"replaced-observed-binding", "PRODUCER_INPUT_MUTATION", sourceReplayReason, false},
		{"extra-unknown-observation", "PRODUCER_INPUT_MUTATION", sourceReplayReason, false},
		{"omitted-declared-subject", "RESEALED_RECEIPT_MUTATION", sourceRebuildReason, true},
		{"replaced-declared-subject", "RESEALED_RECEIPT_MUTATION", sourceRebuildReason, true},
		{"ambiguous-binding", "PRODUCER_INPUT_MUTATION", missingReceiptReason, false},
		{"wrong-head", "RESEALED_RECEIPT_MUTATION", identityReason, false},
		{"resealed-program", "SOURCE_INPUT_MUTATION", programReplayReason, false},
		{"wrong-source-digest", "RESEALED_RECEIPT_MUTATION", sourceRebuildReason, true},
		{"program-source-tamper", "SOURCE_INPUT_MUTATION", programSourceReason, false},
		{"missing-receipt", "MISSING_INPUT", missingReceiptReason, false},
		{"repository-write-authority", "RESEALED_RECEIPT_MUTATION", identityReason, false},
		{"promotion-authority", "RESEALED_RECEIPT_MUTATION", identityReason, false},
		{"wrong-registry-origin", "RESEALED_RECEIPT_MUTATION", evidenceReasonPrefix + "concept-operation registry origin is invalid", false},
		{"missing-producer-input", "MISSING_INPUT", sourceReasonPrefix + "concept-operation strategy source is missing", false},
		{"wrong-source-subject", "SOURCE_INPUT_MUTATION", sourceReasonPrefix + "concept-operation source subject is not exact", false},
	}
}
