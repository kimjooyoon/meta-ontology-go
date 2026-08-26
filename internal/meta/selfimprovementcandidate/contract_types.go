package selfimprovementcandidate

type ContractEvidence struct {
	ContractID      string `json:"contract_id"`
	Path            string `json:"path"`
	Package         string `json:"package"`
	Namespace       string `json:"namespace"`
	EntityCount     int    `json:"entity_count"`
	ActivityCount   int    `json:"activity_count"`
	SourceLines     int    `json:"source_lines"`
	SourceDigest    string `json:"source_digest"`
	CanonicalDigest string `json:"canonical_digest"`
}
