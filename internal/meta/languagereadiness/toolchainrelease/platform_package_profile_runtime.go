package toolchainrelease

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

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
		Deliveries []packageProfileDelivery
	}
}

// A delivery keeps the original positional ports as well as its native value.
type packageProfileDelivery struct {
	ID                      string `json:"activity_id"`
	Input, Actual, Expected json.RawMessage
	Inputs                  []bodyexecution.CompositionPortDelivery `json:"inputs"`
	Passed                  *bool
	Fault                   json.RawMessage `json:"fault"`
	BlockedBy               []string        `json:"blocked_by"`
}
