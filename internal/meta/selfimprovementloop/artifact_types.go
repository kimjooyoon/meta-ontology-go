package selfimprovementloop

type Dossier struct {
	Schema          string         `json:"schema"`
	Scenario        string         `json:"scenario"`
	SourceDigest    string         `json:"source_digest"`
	ToolchainDigest string         `json:"toolchain_digest"`
	Decision        string         `json:"decision"`
	GraphHash       string         `json:"graph_hash"`
	ReportDigest    string         `json:"report_digest"`
	PatchProposal   PatchProposal  `json:"patch_proposal"`
	EvidenceDigest  string         `json:"evidence_digest"`
	Unknowns        []UnknownState `json:"unknowns,omitempty"`
	DossierDigest   string         `json:"dossier_digest"`
}

type Artifacts struct {
	Report        Report
	PatchProposal PatchProposal
	Evidence      EvidenceBundle
	Dossier       Dossier
}
