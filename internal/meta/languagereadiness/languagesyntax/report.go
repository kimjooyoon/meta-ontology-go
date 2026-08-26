package languagesyntax

type Summary struct {
	Satisfied            int `json:"satisfied"`
	Total                int `json:"total"`
	Executed             int `json:"executed"`
	NotSatisfied         int `json:"not_satisfied"`
	Unresolved           int `json:"unresolved"`
	ReadinessBPS         int `json:"readiness_bps"`
	ValidCases           int `json:"valid_cases"`
	InvalidCases         int `json:"invalid_cases"`
	ASTReplays           int `json:"ast_replays"`
	ByteReplays          int `json:"byte_replays"`
	SemanticReplays      int `json:"semantic_replays"`
	GetPutLaws           int `json:"get_put_laws"`
	PutGetLaws           int `json:"put_get_laws"`
	DiagnosticRejections int `json:"diagnostic_rejections"`
	GoooLines            int `json:"gooo_lines"`
	UnregisteredGooo     int `json:"unregistered_gooo"`
	MissingRegistered    int `json:"missing_registered"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema             string       `json:"schema"`
	Decision           string       `json:"decision"`
	Reason             string       `json:"reason"`
	Resolution         string       `json:"resolution"`
	Producer           string       `json:"producer"`
	Consumer           string       `json:"consumer"`
	MetaOperation      string       `json:"meta_operation"`
	HeadSHA            string       `json:"head_sha"`
	Source             Source       `json:"source"`
	Summary            Summary      `json:"summary"`
	Cases              []CaseResult `json:"cases"`
	Indicators         []Indicator  `json:"indicators"`
	Proofs             []Proof      `json:"proofs"`
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthorized bool         `json:"mutation_authorized"`
	ReportDigest       string       `json:"report_digest"`
}

func metric(id, class, proof, resolution string, value, target int) Indicator {
	return Indicator{MetricID: "gooo.metric.language.syntax-roundtrip-" + id + ".v1",
		Class: class, ProofChoice: proof, Producer: "languagesyntax.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: "prove-language-syntax-roundtrip",
		Resolution: resolution, Value: value, Target: target, Satisfied: value == target}
}
