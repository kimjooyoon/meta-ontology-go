package languagesourcebindingpromotion

import "encoding/json"

type producerEnvelope struct {
	Schema            string          `json:"schema"`
	HeadSHA           string          `json:"head_sha"`
	Decision          string          `json:"decision"`
	Resolution        string          `json:"resolution"`
	Reason            string          `json:"reason"`
	ContractDigest    string          `json:"contract_digest"`
	Cases             json.RawMessage `json:"cases"`
	Summary           json.RawMessage `json:"summary"`
	Indicators        json.RawMessage `json:"indicators"`
	Proofs            json.RawMessage `json:"proofs"`
	RepositoryWrites  int             `json:"repository_writes"`
	MutationAuthority bool            `json:"mutation_authority"`
	NotClaimed        json.RawMessage `json:"not_claimed"`
	Digest            string          `json:"digest"`
}

type receiptEnvelope struct {
	Schema         string          `json:"schema"`
	Decision       string          `json:"decision"`
	Reason         string          `json:"reason"`
	Resolution     string          `json:"resolution"`
	Filename       string          `json:"filename"`
	SourceDigest   string          `json:"source_digest"`
	SemanticDigest string          `json:"semantic_digest"`
	Entry          json.RawMessage `json:"entry"`
	Events         json.RawMessage `json:"events"`
	Diagnostics    json.RawMessage `json:"diagnostics"`
	Effects        json.RawMessage `json:"effects"`
	Digest         string          `json:"digest"`
}

type oracleEnvelope struct {
	Schema               string          `json:"schema"`
	Scope                string          `json:"scope"`
	HeadSHA              string          `json:"head_sha"`
	Decision             string          `json:"decision"`
	Resolution           string          `json:"resolution"`
	Reason               string          `json:"reason"`
	ContractDigest       string          `json:"contract_digest"`
	IndependenceDigest   string          `json:"independence_digest"`
	LegacyDigest         string          `json:"legacy_digest"`
	Cases                json.RawMessage `json:"cases"`
	Summary              json.RawMessage `json:"summary"`
	Indicators           json.RawMessage `json:"indicators"`
	NotClaimed           json.RawMessage `json:"not_claimed"`
	RepositoryWrites     int             `json:"repository_writes"`
	MutationAuthority    bool            `json:"mutation_authority"`
	Digest               string          `json:"digest"`
}
