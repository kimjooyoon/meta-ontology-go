package metainvocation

const (
	InputSchema   = "gooo/ci-plan-input/v1"
	PlanSchema    = "gooo/ci-plan/v1"
	ReceiptSchema = "gooo/ci-plan-verification-receipt/v1"
	ReportSchema  = "gooo/meta-invocation-report/v1"

	DecisionPass    = "PASS"
	DecisionClosed  = "FAIL_CLOSED"
	DecisionUnknown = "UNKNOWN"

	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER_RESOLUTION"

	ClaimOpen       = "OPEN"
	ClaimDischarged = "DISCHARGED"
	ClaimRefuted    = "REFUTED"
)

type SourceCoordinate struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column"`
}

type ChangeSet struct {
	Schema string   `json:"schema"`
	CaseID string   `json:"case_id"`
	Files  []string `json:"files"`
}

type RuleEvidence struct {
	ID         string           `json:"id"`
	Operation  string           `json:"operation"`
	File       string           `json:"file"`
	SpecDigest string           `json:"spec_digest"`
	Source     SourceCoordinate `json:"source"`
}

type PlannedCheck struct {
	ID      string         `json:"id"`
	Command string         `json:"command"`
	Files   []string       `json:"files"`
	Reasons []RuleEvidence `json:"reasons"`
}

type CheckPlan struct {
	Schema      string         `json:"schema"`
	CaseID      string         `json:"case_id"`
	InputDigest string         `json:"input_digest"`
	Checks      []PlannedCheck `json:"checks"`
	Digest      string         `json:"digest"`
}

type UnknownCause struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
	File   string `json:"file,omitempty"`
}

type Claim struct {
	ID        string   `json:"id"`
	Statement string   `json:"statement"`
	Status    string   `json:"status"`
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Evidence  []string `json:"evidence"`
	DependsOn []string `json:"depends_on"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type VerificationReceipt struct {
	Schema          string         `json:"schema"`
	SubjectDigest   string         `json:"subject_digest"`
	Decision        string         `json:"decision"`
	Resolution      string         `json:"resolution"`
	EvidenceDigests []string       `json:"evidence_digests"`
	Unknowns        []UnknownCause `json:"unknowns"`
	Digest          string         `json:"digest"`
}

type Report struct {
	Schema       string              `json:"schema"`
	Decision     string              `json:"decision"`
	Resolution   string              `json:"resolution"`
	CaseID       string              `json:"case_id"`
	Entry        string              `json:"entry"`
	SourceDigest string              `json:"source_digest"`
	InputDigest  string              `json:"input_digest"`
	Plan         CheckPlan           `json:"plan"`
	Receipt      VerificationReceipt `json:"receipt"`
	Unknowns     []UnknownCause      `json:"unknowns"`
	Claims       []Claim             `json:"claims"`
	Effects      Effects             `json:"effects"`
	NotClaimed   []string            `json:"not_claimed"`
	ReportDigest string              `json:"report_digest"`
}

type OperationSpec struct {
	ID                string `json:"id"`
	InputEntity       string `json:"input_entity"`
	OutputEntity      string `json:"output_entity"`
	Phase             string `json:"phase"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type BoundOperation struct {
	Activity   string           `json:"activity"`
	Program    string           `json:"program"`
	Input      string           `json:"input"`
	Output     string           `json:"output"`
	SpecDigest string           `json:"spec_digest"`
	Source     SourceCoordinate `json:"source"`
}

type Program struct {
	SourcePath   string                    `json:"source_path"`
	Package      string                    `json:"package"`
	Namespace    string                    `json:"namespace"`
	SourceDigest string                    `json:"source_digest"`
	Entities     map[string]string         `json:"entities"`
	Operations   map[string]BoundOperation `json:"operations"`
}
