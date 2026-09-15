package generation

const IndicatorObservationSchema = "gooo/meta-indicator-observation/v1"

type IndicatorObservation struct {
	Schema              string `json:"schema"`
	IndicatorID         string `json:"indicator_id"`
	Subject             string `json:"subject"`
	HeadSHA             string `json:"head_sha"`
	OperationID         string `json:"operation_id"`
	ValueKind           string `json:"value_kind"`
	ActualValue         int    `json:"actual_value"`
	ExpectedPredicate   string `json:"expected_predicate"`
	ExpectedBound       int    `json:"expected_bound"`
	BeforeFunctionLines int    `json:"before_function_lines"`
	AfterFunctionLines  int    `json:"after_function_lines"`
	TransformedSubject  string `json:"transformed_subject"`
}
