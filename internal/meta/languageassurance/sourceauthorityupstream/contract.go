package sourceauthorityupstream

import "context"

const (
	RequestSchema                 = "gooo/upstream-source-request/v1"
	ReceiptSchema                 = "gooo/upstream-source-receipt/v1"
	SuiteSchema                   = "gooo/upstream-source-conformance/v1"
	DenominatorID                 = "gooo/upstream-source-conformance-denominator/v1"
	ModeExternal                  = "EXTERNAL"
	ObservationSatisfied          = "SATISFIED"
	ObservationUnknown            = "UNKNOWN"
	ResolutionExact               = "EXACT"
	ResolutionInvariantOnly       = "INVARIANT_ONLY"
	EnforcementAllow              = "ALLOW"
	EnforcementBlock              = "BLOCK"
	ReasonSourceSnapshotExact      = "SOURCE_SNAPSHOT_EXACT"
	ReasonPolicyInvalid            = "SOURCE_POLICY_INVALID"
	ReasonRequestInvalid           = "SOURCE_REQUEST_INVALID"
	ReasonAuthorityScopeMismatch   = "AUTHORITY_SCOPE_MISMATCH"
	ReasonFetchFailed              = "SOURCE_FETCH_FAILED"
	ReasonSelectionFailed          = "SOURCE_SELECTION_FAILED"
	ReasonSourceDigestMismatch     = "SOURCE_DIGEST_MISMATCH"
	ReasonSourceSizeMismatch       = "SOURCE_SIZE_MISMATCH"
	ReasonConformanceExact         = "UPSTREAM_CONFORMANCE_EXACT"
	ReasonConformanceCaseMismatch  = "UPSTREAM_CONFORMANCE_CASE_MISMATCH"
	CaseExact                      = "exact"
	CaseDigestMismatch             = "digest-mismatch"
	CaseAuthorityMismatch          = "authority-mismatch"
)

type Authority struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Path       string `json:"path"`
}

type Selection struct {
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`
}

type Policy struct {
	SourceRef      string
	AuthorityRef   string
	URL            string
	Authority      Authority
	Selection      Selection
	ExpectedDigest string
	ExpectedBytes  int
}

type Request struct {
	Schema       string    `json:"schema"`
	SubjectSHA   string    `json:"subject_sha"`
	SourceRef    string    `json:"source_ref"`
	AuthorityRef string    `json:"authority_ref"`
	URL          string    `json:"url"`
	Authority    Authority `json:"authority"`
	Selection    Selection `json:"selection"`
}

type Fetcher interface {
	Fetch(context.Context, string) ([]byte, error)
}
