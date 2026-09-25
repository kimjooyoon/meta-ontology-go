package authorizationfoundation

type suiteDefinition struct {
	ID         string
	Mutation   string
	Decision   string
	Resolution string
}

func suiteDefinitions() []suiteDefinition {
	return []suiteDefinition{
		{"exact", "", "AUTHORIZED_SHADOW", "EXACT"},
		{"metadata-unavailable", "metadata-unavailable", "FAIL_CLOSED", "UNKNOWN"},
		{"prior-unavailable", "prior-unavailable", "FAIL_CLOSED", "UNKNOWN"},
		{"current-unavailable", "current-unavailable", "FAIL_CLOSED", "UNKNOWN"},
		{"metadata-expired", "metadata-expired", "FAIL_CLOSED", "UNKNOWN"},
		{"artifact-id-mismatch", "artifact-id-mismatch", "DENIED", "EXACT"},
		{"archive-digest-mismatch", "archive-digest-mismatch", "DENIED", "EXACT"},
		{"prior-receipt-tamper", "prior-receipt-tamper", "DENIED", "EXACT"},
		{"source-digest-mismatch", "source-digest-mismatch", "DENIED", "EXACT"},
		{"tree-digest-mismatch", "tree-digest-mismatch", "DENIED", "EXACT"},
		{"unknown-reason-mismatch", "unknown-reason-mismatch", "DENIED", "EXACT"},
		{"authority-ceiling", "authority-ceiling", "DENIED", "EXACT"},
	}
}
