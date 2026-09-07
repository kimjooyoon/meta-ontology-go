package main

type costBinding struct {
	Invocation string `json:"invocation_id"`
	Head       string `json:"head_sha"`
	Plan       string `json:"plan_digest"`
	Manifest   string `json:"manifest_digest"`
	Activity   string `json:"activity"`
	Indicator  string `json:"action_indicator_id"`
	Subject    string `json:"subject"`
	Source     string `json:"input_contract_source_digest"`
	Semantic   string `json:"input_contract_semantic_digest"`
	Operation  int    `json:"operation_sequence"`
	Pass       string `json:"pass"`
	Kind       string `json:"command_kind"`
}

type costEvent struct {
	costBinding
	Schema   string `json:"schema"`
	Sequence uint64 `json:"event_sequence"`
	Boundary string `json:"boundary"`
	Cost     *struct {
		State   string `json:"state"`
		Start   uint64 `json:"started_at_event"`
		Elapsed *int64 `json:"elapsed_ns"`
	} `json:"cost"`
}

type costRow struct {
	costBinding
	Start   uint64 `json:"started_at_event"`
	Return  uint64 `json:"returned_at_event"`
	Elapsed int64  `json:"elapsed_ns"`
}

type costReport struct {
	Schema           string    `json:"schema"`
	Scope            string    `json:"scope"`
	Authenticity     string    `json:"source_authenticity"`
	Improvement      string    `json:"improvement"`
	Events           int       `json:"events"`
	UnmeasuredEvents int       `json:"unmeasured_events"`
	UnpairedStarts   int       `json:"unpaired_starts"`
	UnknownReturns   int       `json:"unknown_returns"`
	Rows             []costRow `json:"intervals"`
}

type eventKey struct {
	invocation string
	sequence   uint64
}
