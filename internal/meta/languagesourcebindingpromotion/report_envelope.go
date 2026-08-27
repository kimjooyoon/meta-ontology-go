package languagesourcebindingpromotion

type Report struct {
	Schema               string      `json:"schema"`
	Scope                string      `json:"scope"`
	HeadSHA              string      `json:"head_sha"`
	Decision             string      `json:"decision"`
	Resolution           string      `json:"resolution"`
	Reason               string      `json:"reason"`
	ContractDigest       string      `json:"contract_digest"`
	PolicySourceDigest   string      `json:"policy_source_digest"`
	PolicyArtifactDigest string      `json:"policy_artifact_digest"`
	IndependenceDigest   string      `json:"independence_digest"`
	Cases                []CaseResult `json:"cases"`
	Summary              Summary     `json:"summary"`
	Indicators           []Indicator `json:"indicators"`
	Proofs               []Proof     `json:"proofs"`
	NotClaimed           []string    `json:"not_claimed"`
	RepositoryWrites     int         `json:"repository_writes"`
	MutationAuthority    bool        `json:"mutation_authority"`
	Digest               string      `json:"digest"`
}
