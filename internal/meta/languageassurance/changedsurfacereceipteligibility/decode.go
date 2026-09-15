package changedsurfacereceipteligibility

import "encoding/json"

type assuranceCapsule struct {
	Schema            string                `json:"schema"`
	SubjectSHA        string                `json:"subject_sha"`
	DenominatorDigest string                `json:"denominator_digest"`
	Obligations       []assuranceObligation `json:"obligations"`
	Summary           assuranceSummary      `json:"summary"`
}

type assuranceObligation struct {
	MetricID   string `json:"metric_id"`
	Status     string `json:"status"`
	Resolution string `json:"resolution"`
}

type assuranceSummary struct {
	DenominatorTotal          int `json:"denominator_total"`
	Operating                 int `json:"operating"`
	NotImplemented            int `json:"not_implemented"`
	ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
	UnresolvedIndicators      int `json:"unresolved_indicators"`
	ViolatedGuardrails        int `json:"violated_guardrails"`
	RepositoryWrites          int `json:"repository_writes"`
}

type shadowReportCapsule struct {
	Schema            string            `json:"schema"`
	SubjectSHA        string            `json:"subject_sha"`
	DenominatorID     string            `json:"denominator_id"`
	DenominatorDigest string            `json:"denominator_digest"`
	Decision          string            `json:"decision"`
	Resolution        string            `json:"resolution"`
	EnforcementEffect string            `json:"enforcement_effect"`
	MetaOperations    []shadowOperation `json:"meta_operations"`
	Summary           shadowSummary     `json:"summary"`
	RepositoryWrites  int               `json:"repository_writes"`
}

type shadowOperation struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}

type shadowSummary struct {
	ChangedSurfaces  int `json:"changed_surfaces"`
	ReceiptsObserved int `json:"receipts_observed"`
	BoundReceipts    int `json:"bound_receipts"`
	TotalityBPS      int `json:"totality_bps"`
	UnknownPaths     int `json:"unknown_paths"`
	BlockedPaths     int `json:"blocked_paths"`
}

type shadowSuiteCapsule struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	Decision          string `json:"decision"`
	Resolution        string `json:"resolution"`
	CasesTotal        int    `json:"cases_total"`
	CasesPassed       int    `json:"cases_passed"`
	CoverageBPS       int    `json:"coverage_bps"`
}

func decode[T any](payload []byte) (T, error) {
	var result T
	err := json.Unmarshal(payload, &result)
	return result, err
}
