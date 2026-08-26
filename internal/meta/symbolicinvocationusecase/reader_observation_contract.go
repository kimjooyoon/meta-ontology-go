package symbolicinvocationusecase

const (
	SymbolicReaderRequestResultSchema = "gooo/symbolic-reader-request-result/v1"
	SymbolicReaderRequestResultMetric = "gooo.metric.compiler.symbolic-reader-request-result.v1"

	SymbolicReaderObservationSchema     = "gooo/symbolic-reader-request-user-observation/v1"
	SymbolicReaderObservationMetric     = "gooo.metric.user.symbolic-reader-request-observation.v1"
	SymbolicReaderObservationResolution = "USER_OBSERVATION_ONLY"
	SymbolicReaderObservationTotal      = 10
)

type ReaderObservationCoordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type ReaderObservationEffects struct {
	RepositoryWrites int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type ReaderRequestSelectionInput struct {
	Audience           string `json:"audience"`
	ExpectedResolution string `json:"expected_resolution"`
}

type ReaderRequestViewInput struct {
	Audience            string                       `json:"audience"`
	SourceResolution    string                       `json:"source_resolution"`
	EffectiveResolution string                       `json:"effective_resolution"`
	IndicatorIDs        []string                     `json:"indicator_ids"`
	Coordinates         ReaderObservationCoordinates `json:"coordinates"`
}

type ReaderRequestResultInput struct {
	Schema             string                       `json:"schema"`
	SubjectSHA         string                       `json:"subject_sha"`
	MetricID           string                       `json:"metric_id"`
	Decision           string                       `json:"decision"`
	Resolution         string                       `json:"resolution"`
	Reason             string                       `json:"reason"`
	Request            ReaderRequestSelectionInput  `json:"request"`
	View               ReaderRequestViewInput       `json:"view"`
	Coordinates        ReaderObservationCoordinates `json:"coordinates"`
	Effects            ReaderObservationEffects     `json:"effects"`
	PromotionCreditBPS int                          `json:"promotion_credit_bps"`
	Digest             string                       `json:"digest"`
}
