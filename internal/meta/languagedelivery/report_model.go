package languagedelivery

type Coordinates struct {
	Satisfied      int `json:"satisfied"`
	NotImplemented int `json:"not_implemented"`
	NotSatisfied   int `json:"not_satisfied"`
	Unknown        int `json:"unknown"`
	Total          int `json:"total"`
	BasisPoints    int `json:"basis_points"`
}

type ObligationResult struct {
	ID             string         `json:"id"`
	Audience       Audience       `json:"audience"`
	Class          IndicatorClass `json:"class"`
	MetaOperation  string         `json:"meta_operation"`
	ProofChoice    ProofChoice    `json:"proof_choice"`
	Status         ResultStatus   `json:"status"`
	Reason         string         `json:"reason"`
	Source         SourceName     `json:"source"`
	Observed       int            `json:"observed"`
	Expected       int            `json:"expected"`
	EvidenceDigest string         `json:"evidence_digest"`
}

type SourceObservation struct {
	Source            SourceName `json:"source"`
	Schema            string     `json:"schema"`
	Decision          string     `json:"decision"`
	Resolution        string     `json:"resolution"`
	State             string     `json:"state"`
	Reason            string     `json:"reason"`
	ArtifactID        int64      `json:"artifact_id"`
	ArtifactName      string     `json:"artifact_name"`
	ArchiveDigest     string     `json:"archive_digest"`
	ReportDigest      string     `json:"report_digest"`
	RepositoryWrites  int        `json:"repository_writes"`
	MutationAuthority bool       `json:"mutation_authority"`
}
