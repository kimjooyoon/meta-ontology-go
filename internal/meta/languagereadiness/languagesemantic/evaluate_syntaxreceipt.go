package languagesemantic

type syntaxReceipt struct {
	Schema             string        `json:"schema"`
	Decision           string        `json:"decision"`
	Resolution         string        `json:"resolution"`
	Summary            SyntaxSummary `json:"summary"`
	Source             syntaxSource  `json:"source"`
	Cases              []syntaxCase  `json:"cases"`
	RepositoryWrites   int           `json:"repository_writes"`
	MutationAuthorized bool          `json:"mutation_authorized"`
	ReportDigest       string        `json:"report_digest"`
}
