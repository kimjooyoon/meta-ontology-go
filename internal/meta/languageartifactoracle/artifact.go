package languageartifactoracle

const SourceArtifactSchema = "gooo/source-execution-receipt/v1"

type artifactBinding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type artifactEntry struct {
	Package   string            `json:"package"`
	Namespace string            `json:"namespace"`
	Activity  string            `json:"activity"`
	Inputs    []artifactBinding `json:"inputs"`
	Output    artifactBinding   `json:"output"`
}

type artifactEvent struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
}

type artifactDiagnostic struct {
	Stage   string `json:"stage"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type artifactEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type sourceArtifact struct {
	Schema         string               `json:"schema"`
	Decision       string               `json:"decision"`
	Reason         string               `json:"reason"`
	Resolution     string               `json:"resolution"`
	Filename       string               `json:"filename"`
	SourceDigest   string               `json:"source_digest"`
	SemanticDigest string               `json:"semantic_digest,omitempty"`
	Entry          artifactEntry        `json:"entry"`
	Events         []artifactEvent      `json:"events"`
	Diagnostics    []artifactDiagnostic `json:"diagnostics"`
	Effects        artifactEffects      `json:"effects"`
	Digest         string               `json:"digest"`
}
