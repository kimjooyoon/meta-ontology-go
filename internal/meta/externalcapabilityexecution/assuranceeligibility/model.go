package assuranceeligibility

type ArtifactBinding struct {
	Name           string `json:"name"`
	ObservedDigest string `json:"observed_digest"`
	Exact          bool   `json:"exact"`
}

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

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice    string `json:"choice"`
	Status    string `json:"status"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}
