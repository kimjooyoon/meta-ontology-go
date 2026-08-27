package semanticdeltareceiptconsumer

// Input is intentionally copied from the producer wire boundary. The
// consumer reads the source paths independently before adjudicating.
type Input struct {
	CaseID     string `json:"case_id"`
	SubjectSHA string `json:"subject_sha"`
	BeforePath string `json:"before_path"`
	AfterPath  string `json:"after_path"`
}

type Node struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type Fact struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type Claim struct {
	ID        string `json:"id"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Status    string `json:"status"`
	Stage     string `json:"stage"`
	Step      string `json:"step"`
	Reason    string `json:"reason"`
}

type Snapshot struct {
	SourceDigest     string  `json:"source_digest"`
	Bytes            int     `json:"bytes"`
	Lines            int     `json:"lines"`
	ParseStatus      string  `json:"parse_status"`
	ParseReason      string  `json:"parse_reason"`
	SemanticDigest   string  `json:"semantic_digest"`
	StructuralDigest string  `json:"structural_digest"`
	ClaimDigest      string  `json:"claim_digest"`
	Nodes            []Node  `json:"nodes,omitempty"`
	Facts            []Fact  `json:"facts,omitempty"`
	Claims           []Claim `json:"claims,omitempty"`
}

type TextualDelta struct {
	Changed      bool   `json:"changed"`
	Decision     string `json:"decision"`
	BeforeBytes  int    `json:"before_bytes"`
	AfterBytes   int    `json:"after_bytes"`
	ChangedBytes int    `json:"changed_bytes"`
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
}
