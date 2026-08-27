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
	ID           string  `json:"id"`
	Kind         string  `json:"kind"`
	Namespace    string  `json:"namespace"`
	ValueProgram string  `json:"value_program,omitempty"`
	Fields       []Field `json:"fields,omitempty"`
}

type Field struct {
	ID          string `json:"id"`
	Parent      string `json:"parent"`
	TypeID      string `json:"type_id"`
	Presence    string `json:"presence"`
	Cardinality string `json:"cardinality"`
}

type Fact struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type Claim struct {
	ID                    string `json:"id"`
	ClaimTypeID           string `json:"claim_type_id"`
	Kind                  string `json:"kind"`
	Subject               string `json:"subject"`
	Predicate             string `json:"predicate"`
	Object                string `json:"object"`
	Status                string `json:"status"`
	Stage                 string `json:"stage"`
	Step                  string `json:"step"`
	Reason                string `json:"reason"`
	NormalizedProposition string `json:"normalized_proposition"`
	PropositionDigest     string `json:"proposition_digest"`
	TargetAddress         string `json:"target_address"`
	BeforeSourceDigest    string `json:"before_source_digest,omitempty"`
	AfterSourceDigest     string `json:"after_source_digest,omitempty"`
	BeforeSemanticDigest  string `json:"before_semantic_digest,omitempty"`
	AfterSemanticDigest   string `json:"after_semantic_digest,omitempty"`
	PreservationOf        string `json:"preservation_of,omitempty"`
}

type SemanticComponent struct {
	ID                string `json:"id"`
	Kind              string `json:"kind"`
	Subject           string `json:"subject"`
	Predicate         string `json:"predicate"`
	Object            string `json:"object"`
	PropositionDigest string `json:"proposition_digest"`
}

type ComponentChange struct {
	ID     string            `json:"id"`
	Before SemanticComponent `json:"before"`
	After  SemanticComponent `json:"after"`
}

type SemanticComponentDelta struct {
	Status  string              `json:"status"`
	Added   []SemanticComponent `json:"added,omitempty"`
	Removed []SemanticComponent `json:"removed,omitempty"`
	Changed []ComponentChange   `json:"changed,omitempty"`
}

type ClaimMatch struct {
	Slot        string `json:"slot"`
	BeforeCount int    `json:"before_count"`
	AfterCount  int    `json:"after_count"`
	Reason      string `json:"reason"`
}

type Snapshot struct {
	SourceDigest            string              `json:"source_digest"`
	Bytes                   int                 `json:"bytes"`
	Lines                   int                 `json:"lines"`
	ParseStatus             string              `json:"parse_status"`
	ParseReason             string              `json:"parse_reason"`
	SemanticDigest          string              `json:"semantic_digest"`
	StructuralDigest        string              `json:"structural_digest"`
	ClaimDigest             string              `json:"claim_digest"`
	SemanticComponentDigest string              `json:"semantic_component_digest"`
	Nodes                   []Node              `json:"nodes,omitempty"`
	Facts                   []Fact              `json:"facts,omitempty"`
	Claims                  []Claim             `json:"claims,omitempty"`
	SemanticComponents      []SemanticComponent `json:"semantic_components,omitempty"`
}

type TextualDelta struct {
	Changed                  bool   `json:"changed"`
	Decision                 string `json:"decision"`
	BeforeBytes              int    `json:"before_bytes"`
	AfterBytes               int    `json:"after_bytes"`
	PositionalByteMismatches int    `json:"positional_byte_mismatches"`
	BeforeDigest             string `json:"before_digest"`
	AfterDigest              string `json:"after_digest"`
}
