package languagepackageexecution

type AudienceView struct {
	Audience         string   `json:"audience"`
	ReaderResolution string   `json:"reader_resolution"`
	VisibleFacts     []string `json:"visible_facts"`
	HiddenFacts      int      `json:"hidden_facts"`
	FactsDigest      string   `json:"facts_digest"`
}

type Report struct {
	Schema            string         `json:"schema"`
	Decision          string         `json:"decision"`
	Reason            string         `json:"reason"`
	Resolution        string         `json:"resolution"`
	HeadSHA           string         `json:"head_sha"`
	Summary           Summary        `json:"summary"`
	Cases             []CaseResult   `json:"cases"`
	Indicators        []Indicator    `json:"indicators"`
	Proofs            []Proof        `json:"proofs"`
	Views             []AudienceView `json:"views"`
	FactsDigest       string         `json:"facts_digest"`
	RepositoryWrites  int            `json:"repository_writes"`
	MutationAuthority bool           `json:"mutation_authority"`
	Digest            string         `json:"digest"`
}
