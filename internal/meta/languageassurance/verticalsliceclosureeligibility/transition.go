package verticalsliceclosureeligibility

type Transition struct {
	MetricID           string `json:"metric_id"`
	MetaOperation      string `json:"meta_operation"`
	FromStatus         string `json:"from_status"`
	FromResolution     string `json:"from_resolution"`
	EligibleStatus     string `json:"eligible_status"`
	EligibleResolution string `json:"eligible_resolution"`
	OfficialStatus     string `json:"official_status"`
	OfficialResolution string `json:"official_resolution"`
}

func eligibilityTransition() Transition {
	return Transition{MetricID: MetricID, MetaOperation: MetaOperation,
		FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE",
		EligibleStatus: "OPERATING", EligibleResolution: ResolutionExact,
		OfficialStatus: "NOT_IMPLEMENTED", OfficialResolution: "NONE"}
}
