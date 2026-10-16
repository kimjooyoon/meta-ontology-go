package bodyexecution

import "fmt"

const (
	CompositionInputMaxBytes      = 1 << 20
	CompositionSeriesMaxBytes     = 16 * CompositionInputMaxBytes
	CompositionOriginPageMaxBytes = 32 << 10
	compositionOutputMaxBytes     = 1 << 20
	compositionTextMaxBytes       = 1024
	compositionRowsMax            = 128
)

// CompositionCapacity describes independent aggregate transport limits. Scalar
// and row maxima do not promise acceptance of their full Cartesian product.
// It is an inspection hint; changing a reported value cannot change admission.
type CompositionCapacity struct {
	Schema                      string `json:"schema"`
	InputJSONBytes              int    `json:"input_json_bytes"`
	NormalizedSuiteJSONBytes    int    `json:"normalized_suite_json_bytes"`
	SeriesJSONBytes             int    `json:"series_json_bytes"`
	NativeStdoutBytes           int    `json:"native_stdout_bytes"`
	RecordedOriginPageJSONBytes int    `json:"recorded_origin_page_json_bytes"`
	TextUTF8Bytes               int    `json:"text_utf8_bytes"`
	Rows                        int    `json:"rows"`
	Scope                       string `json:"scope"`
}

func CompositionTransportCapacity() CompositionCapacity {
	return CompositionCapacity{Schema: "gooo/composition-transport-capacity/v1",
		InputJSONBytes: CompositionInputMaxBytes, NormalizedSuiteJSONBytes: CompositionInputMaxBytes,
		SeriesJSONBytes: CompositionSeriesMaxBytes, NativeStdoutBytes: compositionOutputMaxBytes,
		RecordedOriginPageJSONBytes: CompositionOriginPageMaxBytes,
		TextUTF8Bytes:               compositionTextMaxBytes, Rows: compositionRowsMax,
		Scope: "independent aggregate JSON byte budgets; decoded Text bytes and row count are separate; provenance files and saved receipts have their own bounds"}
}

func checkCompositionInputSize(raw []byte, stage string) error {
	if len(raw) > CompositionInputMaxBytes {
		return fmt.Errorf("composition %s JSON is %d bytes; maximum %d bytes", stage, len(raw), CompositionInputMaxBytes)
	}
	return nil
}

func decodeCompositionInputData(raw []byte, output any) error {
	if err := checkCompositionInputSize(raw, "input"); err != nil {
		return err
	}
	return decodeCompositionTextData(raw, output, CompositionInputMaxBytes)
}
