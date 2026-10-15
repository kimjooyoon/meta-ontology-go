package toolchainrelease

import "encoding/json"

type packageProfileRuntime struct {
	Stage, Failure string
	SHA            string `json:"generated_sha256"`
	Passed         *int   `json:"finite_passed"`
	Total          *int   `json:"finite_total"`
	Calls          *int   `json:"model_calls"`
	Projection     *bool  `json:"projection_replayed"`
	Replay         *bool  `json:"runtime_replayed"`
	Traces         []struct {
		Index      *int `json:"case_index"`
		Deliveries []struct {
			ID                      string `json:"activity_id"`
			Input, Actual, Expected json.RawMessage
			Passed                  *bool
		}
	}
}
