package languagetestexperiment

type Report struct {
	Schema            string      `json:"schema"`
	SubjectSHA        string      `json:"subject_sha"`
	Decision          string      `json:"decision"`
	Reason            string      `json:"reason"`
	Resolution        string      `json:"resolution"`
	Summary           Summary     `json:"summary"`
	Indicators        []Indicator `json:"indicators"`
	Views             []View      `json:"views"`
	RepositoryWrites  int         `json:"repository_writes"`
	MutationAuthority bool        `json:"mutation_authority"`
	Digest            string      `json:"digest"`
}
