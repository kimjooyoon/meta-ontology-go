package bodyexecution

const compositionFaultRuntimeSchema = "gooo/body-composition-runtime/v3"

// CompositionFaultSite binds an observed operation to the replayed pure Go
// projection, whose source and typed plan are retained in CompositionRuntime.
type CompositionFaultSite struct {
	ExpressionID     string `json:"expression_id"`
	RootActivityID   string `json:"root_activity_id"`
	ActivityID       string `json:"activity_id"`
	Operator         string `json:"operator"`
	Expression       string `json:"expression"`
	ProjectionSHA256 string `json:"projection_sha256"`
	Start            int    `json:"start"`
	End              int    `json:"end"`
}

type CompositionFault struct {
	Kind  string               `json:"kind"`
	Site  CompositionFaultSite `json:"site"`
	Left  int64                `json:"left"`
	Right int64                `json:"right"`
}

// Outcomes count supplied expectations, not all activity invocations. A fault
// and an unavailable dependency both retain their original expected value.
type CompositionOutcomes struct {
	Matched    int `json:"matched"`
	Mismatched int `json:"mismatched"`
	Faulted    int `json:"faulted"`
	Blocked    int `json:"blocked"`
	Unobserved int `json:"unobserved"`
}

type nativeArithmeticFault struct {
	Site  *int   `json:"site"`
	Left  *int64 `json:"left"`
	Right *int64 `json:"right"`
}

func hasCompositionFault(r CompositionRuntime) bool {
	for _, trace := range r.Traces {
		for _, delivery := range trace.Deliveries {
			if delivery.Fault != nil || len(delivery.BlockedBy) != 0 {
				return true
			}
		}
	}
	return false
}

func completeArithmeticRuns(r CompositionRuntime) bool {
	if len(r.Runs) != 2 || !r.RuntimeReplayed || r.Stage != "COMPLETE" || r.Outcomes == nil {
		return false
	}
	for _, run := range r.Runs {
		if !run.Started || !run.Completed || run.Canceled || run.TimedOut || run.OutputTruncated ||
			run.ExitCode == nil || *run.ExitCode != 0 || run.DiagnosticsBytes != 0 ||
			run.StderrSHA256 != digest(nil) || !runtimeDigest(run.StdoutSHA256) {
			return false
		}
	}
	return r.Runs[0].StdoutSHA256 == r.Runs[1].StdoutSHA256
}
