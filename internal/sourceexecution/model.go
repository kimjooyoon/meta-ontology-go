package sourceexecution

const ReceiptSchema = "gooo/source-execution-receipt/v1"

type Request struct {
	Filename string
	Source   string
	Entry    string
}

type Binding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Entry struct {
	Package   string    `json:"package"`
	Namespace string    `json:"namespace"`
	Activity  string    `json:"activity"`
	Inputs    []Binding `json:"inputs"`
	Output    Binding   `json:"output"`
}

type Event struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
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
	Entry          Entry        `json:"entry"`
	Events         []Event      `json:"events"`
	Diagnostics    []Diagnostic `json:"diagnostics"`
	Effects        Effects      `json:"effects"`
	Digest         string       `json:"digest"`
}
