package languageartifactoracle

const (
	ReportSchema       = "gooo/language-artifact-oracle/v1"
	ReportScope        = "SOURCE_EXECUTION_ARTIFACT_BINDING_ONLY"
	IndependenceSchema = "gooo/language-artifact-oracle-independence/v1"
	CaseTotal          = 4
	CheckTotal         = 9
)

type IndependenceEvidence struct {
	Schema               string `json:"schema"`
	ProducerDependencies int    `json:"producer_dependencies"`
}

type Input struct {
	Contract            Contract
	HeadSHA             string
	Filename            string
	UnsupportedFilename string
	Entry               string
	Source              []byte
	UnsupportedSource   []byte
	Genuine             []byte
	Forged              []byte
	UnknownDecision     []byte
	LegacyAcceptance    []byte
	Independence        IndependenceEvidence
}

type projection struct {
	Package   string
	Namespace string
	Activity  string
	Inputs    []artifactBinding
	Output    artifactBinding
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type CheckResult struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Expected      string `json:"expected"`
	Observed      string `json:"observed"`
}

type oracleResult struct {
	Decision       string
	Resolution     string
	Reason         string
	Coordinate     Coordinate
	Checks         []CheckResult
	SourceDigest   string
	ArtifactDigest string
}
