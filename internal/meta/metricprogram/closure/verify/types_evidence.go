package verify

type rootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	ReadmeRequirement     string `json:"readme_requirement"`
}

type files struct {
	Program      fileEvidence `json:"program"`
	Source       fileEvidence `json:"source"`
	Verification fileEvidence `json:"verification"`
}

type fileEvidence struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type indicator struct {
	ID             string `json:"id"`
	Family         string `json:"family"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}
