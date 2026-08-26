package artifactemit

type SymbolicReaderRequestEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type SymbolicReaderRequestResult struct {
	Schema             string                           `json:"schema"`
	SubjectSHA         string                           `json:"subject_sha"`
	MetricID           string                           `json:"metric_id"`
	Decision           string                           `json:"decision"`
	Resolution         string                           `json:"resolution"`
	Reason             string                           `json:"reason"`
	Request            SymbolicReaderRequestDeclaration `json:"request"`
	Source             SymbolicReaderRequestSource      `json:"source"`
	View               SymbolicReaderRequestView        `json:"view"`
	Coordinates        SymbolicReaderRequestCoordinates `json:"coordinates"`
	Classes            []SymbolicReaderRequestClass     `json:"classes"`
	Indicators         []SymbolicReaderRequestIndicator `json:"indicators"`
	Proofs             []SymbolicReaderRequestProof     `json:"proofs"`
	Effects            SymbolicReaderRequestEffects     `json:"effects"`
	PromotionCreditBPS int                              `json:"promotion_credit_bps"`
	NotClaimed         []string                         `json:"not_claimed"`
	Digest             string                           `json:"digest,omitempty"`
}
