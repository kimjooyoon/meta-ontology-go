package languagetest

const (
	ReceiptSchema = "gooo/language-test-receipt/v1"
	MarkerPrefix  = "gooo://test/activity/"

	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
)

type Request struct {
	Filename string
	Source   string
}

type Binding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Case struct {
	Name            string  `json:"name"`
	MarkerID        string  `json:"marker_id"`
	Entry           string  `json:"entry"`
	Assertion       string  `json:"assertion"`
	Expected        Binding `json:"expected"`
	Observed        Binding `json:"observed"`
	ExecutionDigest string  `json:"execution_digest,omitempty"`
	Decision        string  `json:"decision"`
	Reason          string  `json:"reason"`
}

type Summary struct {
	Declared int `json:"declared"`
	Executed int `json:"executed"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
}

type Diagnostic struct {
	Stage   string `json:"stage"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema         string       `json:"schema"`
	Decision       string       `json:"decision"`
	Reason         string       `json:"reason"`
	Resolution     string       `json:"resolution"`
	Filename       string       `json:"filename"`
	SourceDigest   string       `json:"source_digest"`
	SemanticDigest string       `json:"semantic_digest,omitempty"`
	Summary        Summary      `json:"summary"`
	Cases          []Case       `json:"cases"`
	Diagnostics    []Diagnostic `json:"diagnostics"`
	Effects        Effects      `json:"effects"`
	NonClaims      []string     `json:"non_claims"`
	Digest         string       `json:"digest"`
}

func nonClaims() []string {
	return []string{"value-level-assertions", "side-effect-assertions", "external-dependency-execution"}
}
