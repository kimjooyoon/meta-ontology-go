package sourcepolicy

// Observation is a raw fact produced by a scanner or CI observer.
type Observation struct {
	Subject   string
	Dimension Dimension
	Value     int
	Detail    string
	Producer  string
}

// Indicator joins a metric fact to policy, proof choice, and meta operation.
type Indicator struct {
	MetricID  Dimension   `json:"metric_id"`
	Family    Family      `json:"family"`
	Subject   string      `json:"subject"`
	Value     int         `json:"value"`
	Limit     int         `json:"limit"`
	Relation  Relation    `json:"relation"`
	Blocking  bool        `json:"blocking"`
	Satisfied bool        `json:"satisfied"`
	Proof     ProofChoice `json:"proof_choice"`
	Producer  string      `json:"producer"`
	Consumer  string      `json:"consumer"`
	Operation Operation   `json:"meta_operation"`
	Detail    string      `json:"detail,omitempty"`
}

// Report is the deterministic bridge from metrics to meta operations.
type Report struct {
	Schema     string      `json:"schema"`
	Policy     Policy      `json:"policy"`
	Indicators []Indicator `json:"indicators"`
}

func (r Report) Actionable() []Indicator {
	actionable := make([]Indicator, 0)
	for _, indicator := range r.Indicators {
		if !indicator.Satisfied {
			actionable = append(actionable, indicator)
		}
	}
	return actionable
}

func (r Report) Failed() []Indicator {
	failed := make([]Indicator, 0)
	for _, indicator := range r.Actionable() {
		if indicator.Blocking {
			failed = append(failed, indicator)
		}
	}
	return failed
}

func (r Report) BlockingCount() int {
	count := 0
	for _, indicator := range r.Indicators {
		if indicator.Blocking {
			count++
		}
	}
	return count
}
