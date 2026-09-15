package selfimprovementloop

type Report struct {
	Schema          string            `json:"schema"`
	Scenario        string            `json:"scenario"`
	SourceDigest    string            `json:"source_digest"`
	ToolchainDigest string            `json:"toolchain_digest"`
	Decision        string            `json:"decision"`
	Reason          string            `json:"reason"`
	GraphHash       string            `json:"graph_hash"`
	Cells           []CellResult      `json:"cells"`
	Bindings        []ActivityBinding `json:"bindings"`
	Unknowns        []UnknownState    `json:"unknowns,omitempty"`
	PairMatched     bool              `json:"pair_matched"`
	ReportDigest    string            `json:"report_digest"`
}

type PatchProposal struct {
	Schema             string `json:"schema"`
	Scenario           string `json:"scenario"`
	OutputMode         string `json:"output_mode"`
	RepositoryMutation bool   `json:"repository_mutation"`
	Patch              string `json:"patch"`
}

type EvidenceRecord struct {
	Cell     string        `json:"cell"`
	Decision string        `json:"decision"`
	Reason   string        `json:"reason"`
	Unknown  *UnknownState `json:"unknown,omitempty"`
}

type EvidenceBundle struct {
	Schema          string           `json:"schema"`
	Scenario        string           `json:"scenario"`
	SourceDigest    string           `json:"source_digest"`
	ToolchainDigest string           `json:"toolchain_digest"`
	Decision        string           `json:"decision"`
	Cells           []EvidenceRecord `json:"cells"`
	Unknowns        []UnknownState   `json:"unknowns,omitempty"`
	Pair            ExactPair        `json:"pair"`
	GraphHash       string           `json:"graph_hash"`
	EvidenceDigest  string           `json:"evidence_digest"`
}
