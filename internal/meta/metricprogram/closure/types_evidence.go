package closure

type RootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	ReadmeRequirement     string `json:"readme_requirement"`
}

type Files struct {
	Program      FileEvidence `json:"program"`
	Source       FileEvidence `json:"source"`
	Verification FileEvidence `json:"verification"`
}

type FileEvidence struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type Indicator struct {
	ID             string `json:"id"`
	Family         string `json:"family"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}
