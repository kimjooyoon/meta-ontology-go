package symbolicinvocationusecase

type ReaderObservationSource struct {
	Schema         string `json:"schema"`
	MetricID       string `json:"metric_id"`
	SubjectSHA     string `json:"subject_sha"`
	Decision       string `json:"decision"`
	Resolution     string `json:"resolution"`
	ArtifactDigest string `json:"artifact_digest"`
	FileDigest     string `json:"file_digest"`
}

type ReaderObservationView struct {
	Audience             string   `json:"audience"`
	EffectiveResolution  string   `json:"effective_resolution"`
	SelectedIndicatorIDs []string `json:"selected_indicator_ids"`
}

type ReaderObservationIndicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type ReaderObservationClassCoordinates struct {
	Class     string `json:"class"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}

type ReaderObservationProofCoordinates struct {
	ProofChoice string `json:"proof_choice"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
}

type ReaderObservationReport struct {
	Schema             string                              `json:"schema"`
	SubjectSHA         string                              `json:"subject_sha"`
	MetricID           string                              `json:"metric_id"`
	Decision           string                              `json:"decision"`
	Resolution         string                              `json:"resolution"`
	Reason             string                              `json:"reason"`
	Source             ReaderObservationSource             `json:"source"`
	View               ReaderObservationView               `json:"view"`
	Coordinates        ReaderObservationCoordinates        `json:"coordinates"`
	Classes            []ReaderObservationClassCoordinates `json:"classes"`
	Indicators         []ReaderObservationIndicator        `json:"indicators"`
	Proofs             []ReaderObservationProofCoordinates `json:"proofs"`
	Effects            ReaderObservationEffects            `json:"effects"`
	PromotionCreditBPS int                                 `json:"promotion_credit_bps"`
	NotClaimed         []string                            `json:"not_claimed"`
	Digest             string                              `json:"digest"`
}
