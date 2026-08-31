package artifactemit

const (
	symbolicReaderRequestSchema = "gooo/symbolic-reader-request-result/v1"
	symbolicReaderRequestMetric = "gooo.metric.compiler.symbolic-reader-request-result.v1"
)

type SymbolicReaderRequestCoordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type SymbolicReaderRequestClass struct {
	Class     string `json:"class"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}

type SymbolicReaderRequestIndicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type SymbolicReaderRequestProof struct {
	ProofChoice string `json:"proof_choice"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
}

type SymbolicReaderRequestDeclaration struct {
	Package            string `json:"package"`
	Namespace          string `json:"namespace"`
	Activity           string `json:"activity"`
	ValueProgram       string `json:"value_program"`
	Audience           string `json:"audience"`
	ExpectedResolution string `json:"expected_resolution"`
	SourceDigest       string `json:"source_digest"`
}

type SymbolicReaderRequestSource struct {
	Schema     string `json:"schema"`
	MetricID   string `json:"metric_id"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Digest     string `json:"digest"`
	FileDigest string `json:"file_digest"`
}

type SymbolicReaderRequestView struct {
	Audience            string                           `json:"audience"`
	SourceResolution    string                           `json:"source_resolution"`
	EffectiveResolution string                           `json:"effective_resolution"`
	IndicatorIDs        []string                         `json:"indicator_ids"`
	Coordinates         SymbolicValueContractCoordinates `json:"coordinates"`
}
