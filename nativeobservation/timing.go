package nativeobservation

// ProcessTiming separates the host Start call from Wait, including pipe I/O.
// Neither interval measures when generated user code began running.
type ProcessTiming struct {
	StartNS              int64  `json:"start_ns"`
	WaitNS               int64  `json:"wait_ns"`
	WaitLimitNS          int64  `json:"wait_limit_ns,omitempty"`
	DeadlineRemainingNS  *int64 `json:"deadline_remaining_ns,omitempty"`
	ContextAtStartReturn string `json:"context_at_start_return"`
}
