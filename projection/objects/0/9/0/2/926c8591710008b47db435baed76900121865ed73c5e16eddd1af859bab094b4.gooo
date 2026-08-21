package bindingcoverage

type PrecedenceEntry struct {
	Rank   uint64 `json:"rank"`
	Stage  string `json:"stage"`
	Reason string `json:"reason"`
}
type Partition struct {
	PartitionID    string   `json:"partition_id"`
	BindingID      string   `json:"binding_id"`
	Polarity       Polarity `json:"polarity"`
	ExpectedStage  string   `json:"expected_stage"`
	ExpectedReason string   `json:"expected_reason"`
}
type Input struct {
	SchemaVersion          string            `json:"schema_version"`
	ContractID             string            `json:"contract_id"`
	SnapshotDigest         string            `json:"snapshot_digest"`
	ExpectedSnapshotDigest string            `json:"expected_snapshot_digest"`
	RequiredBindings       []RequiredBinding `json:"required_bindings"`
	Partitions             []Partition       `json:"partitions"`
	PrecedenceRegistry     []PrecedenceEntry `json:"precedence_registry"`
}
type Output struct {
	SchemaVersion             string   `json:"schema_version"`
	ContractID                string   `json:"contract_id"`
	SnapshotDigest            string   `json:"snapshot_digest"`
	ExpectedSnapshotDigest    string   `json:"expected_snapshot_digest"`
	InputDigest               string   `json:"input_digest"`
	RequiredBindingCount      uint64   `json:"required_binding_count"`
	MatchCoveredCount         uint64   `json:"match_covered_count"`
	MismatchCoveredCount      uint64   `json:"mismatch_covered_count"`
	PartitionCount            uint64   `json:"partition_count"`
	EndpointReferenceCount    uint64   `json:"endpoint_reference_count"`
	DeterministicWorkUnits    uint64   `json:"deterministic_work_units"`
	InputBytes                uint64   `json:"input_bytes"`
	Decision                  Decision `json:"decision"`
	Reason                    Reason   `json:"reason"`
	MissingMatchBindingIDs    []string `json:"missing_match_binding_ids"`
	MissingMismatchBindingIDs []string `json:"missing_mismatch_binding_ids"`
	CanonicalDigest           string   `json:"canonical_digest"`
}
