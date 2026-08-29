package main

type splitEvidence struct {
	OperationID      string                 `json:"operation_id"`
	ExpectedHeadSHA  string                 `json:"expected_head_sha"`
	EvidenceComplete bool                   `json:"evidence_complete"`
	Source           splitEvidenceFile      `json:"source"`
	Candidates       []splitEvidenceFile    `json:"candidates"`
	BuildContexts    []splitEvidenceContext `json:"build_contexts"`
	Write            splitWriteEvidence     `json:"write_receipt"`
}

type splitEvidenceFile struct {
	Path             string   `json:"path"`
	Data             []byte   `json:"data"`
	DeclarationOrder []string `json:"declaration_order,omitempty"`
}

type splitEvidenceContext struct {
	GOOS       string   `json:"goos"`
	GOARCH     string   `json:"goarch"`
	CgoEnabled bool     `json:"cgo_enabled"`
	BuildTags  []string `json:"build_tags"`
}

type splitWriteEvidence struct {
	Complete                     bool                 `json:"complete"`
	ExecutionSucceeded           bool                 `json:"execution_succeeded"`
	DeclaredTargets              []string             `json:"declared_targets"`
	Events                       []splitEvidenceEvent `json:"events"`
	WritesOutsideDeclaredTargets int                  `json:"writes_outside_declared_targets"`
	TemporaryFilesRemaining      int                  `json:"temporary_files_remaining"`
}

type splitEvidenceEvent struct {
	Kind      string `json:"kind"`
	Target    string `json:"target"`
	Temporary string `json:"temporary"`
	Success   bool   `json:"success"`
}
