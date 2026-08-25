package main

type splitBatchCoordinates struct {
	SelectedSubjects int `json:"selected_subjects"`
	AppliedSubjects  int `json:"applied_subjects"`
	ChangedPaths     int `json:"changed_paths"`
	CreatedPaths     int `json:"created_paths"`
	DeferredTopology int `json:"deferred_topology_subjects"`
	Unknowns         int `json:"unknowns"`
}

type splitBatchIndicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type splitBatchProof struct {
	Choice        string `json:"choice"`
	MetaOperation string `json:"meta_operation"`
	Passed        bool   `json:"passed"`
}

type splitBatchReport struct {
	Schema        string                `json:"schema"`
	SourceSHA     string                `json:"source_sha"`
	Decision      string                `json:"decision"`
	Resolution    string                `json:"resolution"`
	Reason        string                `json:"reason"`
	MetaOperation string                `json:"meta_operation"`
	Selected      []string              `json:"selected"`
	Unhandled     []string              `json:"unhandled"`
	Subjects      []splitBatchSubject   `json:"subjects"`
	Coordinates   splitBatchCoordinates `json:"coordinates"`
	Indicators    []splitBatchIndicator `json:"indicators"`
	Proofs        []splitBatchProof     `json:"proofs"`
	SuccessorProjectionRequired bool    `json:"successor_projection_required"`
	Exact         bool                  `json:"exact"`
	Digest        string                `json:"digest"`
}
