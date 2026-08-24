package verticalsliceclosureshadow

type Input struct {
	HeadSHA   string
	Assurance []byte
	Artifacts map[string][]byte
}

type denominator struct {
	Schema        string         `json:"schema"`
	DenominatorID string         `json:"denominator_id"`
	Version       int            `json:"version"`
	Boundaries    []boundarySpec `json:"boundaries"`
}

type boundarySpec struct {
	ID            string `json:"id"`
	Schema        string `json:"schema"`
	MetaOperation string `json:"meta_operation"`
	Target        int    `json:"target"`
	LinkTarget    int    `json:"link_target"`
}

type artifactEnvelope struct {
	Schema             string            `json:"schema"`
	Decision           string            `json:"decision"`
	Resolution         string            `json:"resolution"`
	HeadSHA            string            `json:"head_sha"`
	MetaOperation      string            `json:"meta_operation"`
	ReportDigest       string            `json:"report_digest"`
	RepositoryWrites   int               `json:"repository_writes"`
	MutationAuthorized *bool             `json:"mutation_authorized"`
	Source             artifactSource    `json:"source"`
	Summary            artifactSummary   `json:"summary"`
	Surfaces           []artifactSurface `json:"surfaces"`
	Cases              []artifactCase    `json:"cases"`
}

type artifactSource struct {
	ExpectedHeadSHA       string `json:"expected_head_sha"`
	MetaOperation         string `json:"meta_operation"`
	ConceptArtifactDigest string `json:"concept_artifact_digest"`
	SyntaxArtifactDigest  string `json:"syntax_artifact_digest"`
	SyntaxReportDigest    string `json:"syntax_report_digest"`
	SemanticFileDigest    string `json:"semantic_file_digest"`
	SemanticReportDigest  string `json:"semantic_report_digest"`
}

type artifactSurface struct {
	ID      string `json:"id"`
	Schema  string `json:"schema"`
	Status  string `json:"status"`
	HeadSHA string `json:"head_sha"`
	Cases   int    `json:"cases"`
}

type artifactCase struct {
	ID       string `json:"id"`
	TargetID string `json:"target_id"`
	Observed string `json:"observed"`
	Expected string `json:"expected"`
}
