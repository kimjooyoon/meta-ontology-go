package authorization

type UnknownEvidence struct {
	Stage       string `json:"stage"`
	IndicatorID string `json:"indicator_id"`
	Reason      string `json:"reason"`
}
