package languagedebug

import "encoding/json"

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionFailClosed Decision = "FAIL_CLOSED"
	ResolutionExact             = "EXACT"
	ResolutionLower             = "LOWER_RESOLUTION"
	StatePaused                 = "PAUSED"
	StateRejected               = "REJECTED"
)

type Event struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema            string            `json:"schema"`
	Decision          Decision          `json:"decision"`
	Reason            string            `json:"reason"`
	Resolution        string            `json:"resolution"`
	State             string            `json:"state"`
	Filename          string            `json:"filename"`
	SourceDigest      string            `json:"source_digest"`
	SemanticDigest    string            `json:"semantic_digest"`
	ExecutionDigest   string            `json:"execution_digest"`
	Entry             json.RawMessage   `json:"entry"`
	Breakpoint        string            `json:"breakpoint"`
	CurrentEvent      *Event            `json:"current_event,omitempty"`
	Trace             []Event           `json:"trace"`
	RemainingEvents   int               `json:"remaining_events"`
	Diagnostics       []json.RawMessage `json:"diagnostics"`
	Effects           Effects           `json:"effects"`
	NonClaims         []string          `json:"non_claims"`
	Digest            string            `json:"digest"`
}

func CanonicalNonClaims() []string {
	return []string{"interactive-control", "time-travel", "variable-inspection"}
}
