package bodyexecution

// CompositionConditionSite locates an if condition in the retained pure Go
// projection. Byte offsets refer to ProjectionSHA256, before instrumentation.
// The activity ID remains the source-declared identity across display renames.
type CompositionConditionSite struct {
	ExpressionID     string `json:"expression_id"`
	RootActivityID   string `json:"root_activity_id"`
	ActivityID       string `json:"activity_id"`
	Expression       string `json:"expression"`
	ProjectionSHA256 string `json:"projection_sha256"`
	Start            int    `json:"start"`
	End              int    `json:"end"`
}

// A compact event references the runtime's site table. Repeated calls and loop
// iterations retain separate events in actual evaluation order.
type CompositionConditionObservation struct {
	SiteIndex int  `json:"site_index"`
	Value     bool `json:"value"`
}

const conditionObservationSchema = "gooo/native-condition-observation/v1"
const conditionObservationLimit = 65536

const conditionObservationFile = "gooo-native-condition-observation.json"

// Summary keeps successful value execution separate from partial diagnostic
// capture. Dropped events retain a per-case count on the corresponding trace.
type CompositionConditionSummary struct {
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	RecordedEvents int    `json:"recorded_events"`
	DroppedEvents  int    `json:"dropped_events"`
	SHA256         string `json:"sha256"`
}
