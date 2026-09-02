package selfimprovementloop

// UnknownState is deliberately not compressed into a message or score.
type UnknownState struct {
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

type ActivityBinding struct {
	Cell       string `json:"cell"`
	Activity   string `json:"activity"`
	ActivityID string `json:"activity_id"`
}

type CellResult struct {
	Cell       string        `json:"cell"`
	Activity   string        `json:"activity"`
	ActivityID string        `json:"activity_id"`
	Decision   string        `json:"decision"`
	Reason     string        `json:"reason"`
	Unknown    *UnknownState `json:"unknown,omitempty"`
}
