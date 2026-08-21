package bindingcoverage

type productionReceipt struct {
	CaseName                string   `json:"case_name"`
	Decision                string   `json:"decision"`
	Reason                  string   `json:"reason"`
	RequiredBindingCount    uint64   `json:"required_binding_count"`
	PartitionCount          uint64   `json:"partition_count"`
	EndpointReferenceCount  uint64   `json:"endpoint_reference_count"`
	WorkUnits               uint64   `json:"work_units"`
	MissingMatch            []string `json:"missing_match"`
	MissingMismatch         []string `json:"missing_mismatch"`
	ProductionInputBytes    uint64   `json:"production_input_bytes"`
	ProductionInputDigest   string   `json:"production_input_digest"`
	ProductionOutputDigest  string   `json:"production_output_digest"`
	AuthorizationFieldsGone bool     `json:"authorization_fields_absent"`
}

const fixedProductionContractID = "urn:bindingcoverage:contract/selective-ci"

type productionInputWire struct {
	SchemaVersion          string                     `json:"schema_version"`
	ContractID             string                     `json:"contract_id"`
	SnapshotDigest         string                     `json:"snapshot_digest"`
	ExpectedSnapshotDigest string                     `json:"expected_snapshot_digest"`
	RequiredBindings       []productionBindingWire    `json:"required_bindings"`
	Partitions             []productionPartitionWire  `json:"partitions"`
	PrecedenceRegistry     []productionPrecedenceWire `json:"precedence_registry"`
}
type productionBindingWire struct {
	BindingID      string `json:"binding_id"`
	FromFieldID    string `json:"from_field_id"`
	ToFieldID      string `json:"to_field_id"`
	Kind           string `json:"kind"`
	ExpectedStage  string `json:"expected_stage"`
	ExpectedReason string `json:"expected_reason"`
}
type productionPartitionWire struct {
	PartitionID    string `json:"partition_id"`
	BindingID      string `json:"binding_id"`
	Polarity       string `json:"polarity"`
	ExpectedStage  string `json:"expected_stage"`
	ExpectedReason string `json:"expected_reason"`
}
type productionPrecedenceWire struct {
	Rank   uint64 `json:"rank"`
	Stage  string `json:"stage"`
	Reason string `json:"reason"`
}
