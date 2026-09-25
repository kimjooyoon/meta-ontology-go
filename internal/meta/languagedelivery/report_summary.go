package languagedelivery

type NamedCoordinates struct {
	Name        string      `json:"name"`
	Coordinates Coordinates `json:"coordinates"`
}

type InternalReadiness struct {
	Claim       string `json:"claim"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
}

type EffectSummary struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Summary struct {
	Coordinates         Coordinates        `json:"coordinates"`
	ByClass             []NamedCoordinates `json:"by_class"`
	ByOwner             []NamedCoordinates `json:"by_owner"`
	MetaBindings        int                `json:"meta_bindings"`
	MetaBindingsTotal   int                `json:"meta_bindings_total"`
	SourceReceipts      int                `json:"source_receipts"`
	SourceReceiptsTotal int                `json:"source_receipts_total"`
	SelfMintedCredits   int                `json:"self_minted_credits"`
	InternalReadiness   InternalReadiness  `json:"internal_readiness"`
	Effects             EffectSummary      `json:"effects"`
}

type AudienceView struct {
	Audience             Audience    `json:"audience"`
	ProjectionDecision   string      `json:"projection_decision"`
	ProjectionResolution string      `json:"projection_resolution"`
	ReceiptDecision      string      `json:"receipt_decision"`
	ReceiptResolution    string      `json:"receipt_resolution"`
	Coordinates          Coordinates `json:"coordinates"`
	VisibleUnknowns      int         `json:"visible_unknowns"`
	HiddenUnknowns       int         `json:"hidden_unknowns"`
	EvidenceDigest       string      `json:"evidence_digest"`
}
