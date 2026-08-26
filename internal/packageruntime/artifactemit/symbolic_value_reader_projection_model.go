package artifactemit

const (
	symbolicReaderProjectionSchema = "gooo/symbolic-value-reader-projection/v1"
	symbolicReaderProjectionMetric = "gooo.metric.compiler.symbolic-value-reader-projection.v1"
)

type SymbolicValueReaderProjectionSource struct {
	Schema             string `json:"schema"`
	MetricID           string `json:"metric_id"`
	Decision           string `json:"decision"`
	Resolution         string `json:"resolution"`
	ReachabilityDigest string `json:"reachability_digest"`
	FileDigest         string `json:"file_digest"`
}

type SymbolicValueReaderProjectionView struct {
	Audience            string                           `json:"audience"`
	SourceResolution    string                           `json:"source_resolution"`
	EffectiveResolution string                           `json:"effective_resolution"`
	IndicatorIDs        []string                         `json:"indicator_ids"`
	Coordinates         SymbolicValueContractCoordinates `json:"coordinates"`
}

type SymbolicValueReaderProjection struct {
	Schema             string                                `json:"schema"`
	SubjectSHA         string                                `json:"subject_sha"`
	MetricID           string                                `json:"metric_id"`
	Decision           string                                `json:"decision"`
	Resolution         string                                `json:"resolution"`
	Reason             string                                `json:"reason"`
	Source             SymbolicValueReaderProjectionSource   `json:"source"`
	Readers            []SymbolicValueReaderProjectionView   `json:"readers"`
	Coordinates        SymbolicValueContractCoordinates      `json:"coordinates"`
	Classes            []SymbolicValueContractClass          `json:"classes"`
	Indicators         []SymbolicValueContractIndicator      `json:"indicators"`
	Views              []SymbolicValueContractView           `json:"views"`
	Proofs             []SymbolicValueContractProof          `json:"proofs"`
	Effects            SymbolicValueContractEffects          `json:"effects"`
	PromotionCreditBPS int                                   `json:"promotion_credit_bps"`
	NotClaimed         []string                              `json:"not_claimed"`
	Digest             string                                `json:"digest,omitempty"`
}
