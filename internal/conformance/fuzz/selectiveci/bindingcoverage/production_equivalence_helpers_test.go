package bindingcoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
)

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

func translateInput(input Input) (production.Input, error) {
	result := production.Input{
		SchemaVersion:          translateSchema(input.Schema),
		ContractID:             fixedProductionContractID,
		SnapshotDigest:         translateDigest(input.SnapshotDigest),
		ExpectedSnapshotDigest: translateDigest(input.ExpectedDigest),
	}
	if input.Precedence != nil {
		result.PrecedenceRegistry = make([]production.PrecedenceEntry, 0, len(input.Precedence))
	}
	if input.RequiredBindings != nil {
		result.RequiredBindings = make([]production.RequiredBinding, 0, len(input.RequiredBindings))
	}
	if input.Partitions != nil {
		result.Partitions = make([]production.Partition, 0, len(input.Partitions))
	}
	for _, entry := range input.Precedence {
		if entry.Rank < 0 {
			return production.Input{}, fmt.Errorf("negative precedence rank")
		}
		result.PrecedenceRegistry = append(result.PrecedenceRegistry, production.PrecedenceEntry{
			Rank: uint64(entry.Rank), Stage: entry.Stage, Reason: entry.Reason,
		})
	}
	for _, binding := range input.RequiredBindings {
		result.RequiredBindings = append(result.RequiredBindings, production.RequiredBinding{
			BindingID: translateID(binding.ID), FromFieldID: translateID(binding.FromFieldID),
			ToFieldID: translateID(binding.ToFieldID), Kind: production.BindingKind(binding.Kind),
			ExpectedStage: binding.ExpectedStage, ExpectedReason: binding.ExpectedReason,
		})
	}
	partitionOrdinals := make(map[string]int)
	for _, partition := range input.Partitions {
		key := partition.BindingID + "\x00" + partition.Polarity
		ordinal := partitionOrdinals[key]
		partitionOrdinals[key] = ordinal + 1
		result.Partitions = append(result.Partitions, production.Partition{
			PartitionID: generatedPartitionID(partition, ordinal),
			BindingID:   translateID(partition.BindingID), Polarity: production.Polarity(partition.Polarity),
			ExpectedStage: partition.Stage, ExpectedReason: partition.Reason,
		})
	}
	return result, nil
}

func translateSchema(schema string) string {
	if schema == SchemaV1 {
		return production.SchemaVersion
	}
	if strings.HasPrefix(schema, "binding-coverage/") {
		return "gooo/selective-ci-binding-coverage/" + strings.TrimPrefix(schema, "binding-coverage/")
	}
	return schema
}

func translateDigest(digest string) string {
	if strings.HasPrefix(digest, "sha256:") {
		return strings.TrimPrefix(digest, "sha256:")
	}
	return digest
}

func translateID(id string) string {
	if strings.HasPrefix(id, "sid:") {
		return "urn:bindingcoverage:" + strings.TrimPrefix(id, "sid:")
	}
	return id
}

func generatedPartitionID(partition Partition, ordinal int) string {
	seed := partition.BindingID + "\x00" + partition.Polarity + "\x00" + partition.Stage + "\x00" + partition.Reason
	sum := sha256.Sum256([]byte(seed))
	return "urn:bindingcoverage:partition/" + hex.EncodeToString(sum[:8]) + "/" + strconv.Itoa(ordinal)
}

func expectedProductionReason(name, oracleReason string) (production.Reason, error) {
	switch name {
	case "selective-ci-9-binding-complete":
		return requireReason(oracleReason, "COMPLETE", production.ReasonComplete)
	case "selective-ci-9-missing-lane-registry-mismatch":
		return requireReason(oracleReason, "MISSING_MISMATCH", production.ReasonMissingMismatch)
	case "missing-match":
		return requireReason(oracleReason, "MISSING_MATCH", production.ReasonMissingMatch)
	case "zero-denominator":
		return requireReason(oracleReason, "ZERO_DENOMINATOR", production.ReasonZeroDenominator)
	case "duplicate-binding-id":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonDuplicateID)
	case "duplicate-binding-polarity":
		return requireReason(oracleReason, "DUPLICATE_PARTITION_POLARITY", production.ReasonDuplicatePolarity)
	case "dangling-binding":
		return requireReason(oracleReason, "DANGLING_PARTITION", production.ReasonUnknownReference)
	case "bad-stable-id":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonInvalidID)
	case "bad-snapshot-digest":
		return requireReason(oracleReason, "STALE_OR_BAD_DIGEST", production.ReasonInvalidDigest)
	case "unsupported-kind":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonInvalidEnum)
	case "bad-token":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonInvalidToken)
	case "stale-partition-metadata":
		return requireReason(oracleReason, "STALE_PARTITION", production.ReasonStalePartition)
	case "unknown-schema":
		return requireReason(oracleReason, "UNKNOWN_SCHEMA", production.ReasonUnknownSchema)
	case "complete-two-bindings", "permuted-complete-two-bindings", "overflow-boundary", "precedence-metadata-as-data", "shared-endpoint-references":
		return requireReason(oracleReason, "COMPLETE", production.ReasonComplete)
	case "valid-but-unequal-snapshot-digests":
		return requireReason(oracleReason, "STALE_OR_BAD_DIGEST", production.ReasonSnapshotMismatch)
	case "self-link-binding":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonSelfLink)
	default:
		return "", fmt.Errorf("unregistered corpus case %q", name)
	}
}

func requireReason(actual, expected string, productionReason production.Reason) (production.Reason, error) {
	if actual != expected {
		return "", fmt.Errorf("oracle reason %q, expected %q", actual, expected)
	}
	return productionReason, nil
}

func productionInputReceipt(input production.Input, output production.Output) (uint64, string, error) {
	canonical, err := marshalProductionInput(input)
	if err != nil {
		return 0, "", err
	}
	sum := sha256.Sum256(canonical)
	digest := hex.EncodeToString(sum[:])
	if output.InputBytes != uint64(len(canonical)) || output.InputDigest != digest {
		return uint64(len(canonical)), digest, fmt.Errorf("production input receipt mismatch: output=%d/%s recomputed=%d/%s", output.InputBytes, output.InputDigest, len(canonical), digest)
	}
	return uint64(len(canonical)), digest, nil
}

func marshalProductionInput(input production.Input) ([]byte, error) {
	wire := productionInputWire{
		SchemaVersion: input.SchemaVersion, ContractID: input.ContractID,
		SnapshotDigest: input.SnapshotDigest, ExpectedSnapshotDigest: input.ExpectedSnapshotDigest,
	}
	if input.RequiredBindings != nil {
		wire.RequiredBindings = make([]productionBindingWire, 0, len(input.RequiredBindings))
	}
	if input.Partitions != nil {
		wire.Partitions = make([]productionPartitionWire, 0, len(input.Partitions))
	}
	if input.PrecedenceRegistry != nil {
		wire.PrecedenceRegistry = make([]productionPrecedenceWire, 0, len(input.PrecedenceRegistry))
	}
	for _, binding := range input.RequiredBindings {
		wire.RequiredBindings = append(wire.RequiredBindings, productionBindingWire{
			BindingID: binding.BindingID, FromFieldID: binding.FromFieldID, ToFieldID: binding.ToFieldID,
			Kind: string(binding.Kind), ExpectedStage: binding.ExpectedStage, ExpectedReason: binding.ExpectedReason,
		})
	}
	for _, partition := range input.Partitions {
		wire.Partitions = append(wire.Partitions, productionPartitionWire{
			PartitionID: partition.PartitionID, BindingID: partition.BindingID, Polarity: string(partition.Polarity),
			ExpectedStage: partition.ExpectedStage, ExpectedReason: partition.ExpectedReason,
		})
	}
	for _, entry := range input.PrecedenceRegistry {
		wire.PrecedenceRegistry = append(wire.PrecedenceRegistry, productionPrecedenceWire{
			Rank: entry.Rank, Stage: entry.Stage, Reason: entry.Reason,
		})
	}
	sort.SliceStable(wire.RequiredBindings, func(i, j int) bool {
		return productionBindingKey(wire.RequiredBindings[i]) < productionBindingKey(wire.RequiredBindings[j])
	})
	sort.SliceStable(wire.Partitions, func(i, j int) bool {
		return productionPartitionKey(wire.Partitions[i]) < productionPartitionKey(wire.Partitions[j])
	})
	sort.SliceStable(wire.PrecedenceRegistry, func(i, j int) bool {
		return productionPrecedenceKey(wire.PrecedenceRegistry[i]) < productionPrecedenceKey(wire.PrecedenceRegistry[j])
	})
	return json.Marshal(wire)
}

func productionBindingKey(binding productionBindingWire) string {
	return binding.BindingID + "\x00" + binding.FromFieldID + "\x00" + binding.ToFieldID + "\x00" + binding.Kind + "\x00" + binding.ExpectedStage + "\x00" + binding.ExpectedReason
}

func productionPartitionKey(partition productionPartitionWire) string {
	return partition.PartitionID + "\x00" + partition.BindingID + "\x00" + partition.Polarity + "\x00" + partition.ExpectedStage + "\x00" + partition.ExpectedReason
}

func productionPrecedenceKey(entry productionPrecedenceWire) string {
	return fmt.Sprintf("%020d\x00%s\x00%s", entry.Rank, entry.Stage, entry.Reason)
}

func authorizationFieldsAbsent(output production.Output) (bool, error) {
	data, err := json.Marshal(output)
	if err != nil {
		return false, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return false, err
	}
	for _, field := range []string{"execution_authorized", "ci_authorized"} {
		if _, exists := fields[field]; exists {
			return false, fmt.Errorf("production output contains %q", field)
		}
	}
	return true, nil
}

func mapProductionIDs(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !strings.HasPrefix(value, "urn:bindingcoverage:") {
			return nil, fmt.Errorf("unexpected production ID %q", value)
		}
		result = append(result, "sid:"+strings.TrimPrefix(value, "urn:bindingcoverage:"))
	}
	sort.Strings(result)
	return result, nil
}

func receiptDigest(receipts []productionReceipt) (string, error) {
	ordered := append([]productionReceipt{}, receipts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CaseName < ordered[j].CaseName })
	data, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
