package pressureindependence

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

const fixtureID = "cycle-7-pressure-independence"

// These bytes are synthetic fixture authorities. They are not live repository
// snapshots, policy, registries, oracle state, or toolchain state.
//
//go:embed authority_snapshot.synthetic.json
var authoritySnapshotArtifactBytes []byte

//go:embed policy.json
var policyArtifactBytes []byte

//go:embed pressure_registry.json
var pressureRegistryArtifactBytes []byte

//go:embed oracle_contract.json
var oracleContractArtifactBytes []byte

//go:embed toolchain_options.json
var toolchainOptionsArtifactBytes []byte

type authorityArtifact struct {
	Schema      string `json:"schema"`
	FixtureID   string `json:"fixture_id"`
	InputSchema string `json:"input_schema"`
	Authority   string `json:"authority_kind"`
	SourceID    string `json:"source_id"`
	SemanticID  string `json:"semantic_id"`
}

type policyArtifact struct {
	Schema             string `json:"schema"`
	FixtureID          string `json:"fixture_id"`
	InputSchema        string `json:"input_schema"`
	PolicyKind         string `json:"policy_kind"`
	RequestedK         uint64 `json:"requested_K"`
	MinimumIndependent uint64 `json:"minimum_independent"`
}

type registryArtifact struct {
	Schema      string           `json:"schema"`
	FixtureID   string           `json:"fixture_id"`
	InputSchema string           `json:"input_schema"`
	Records     []PressureRecord `json:"records"`
}

type oracleArtifact struct {
	Schema                 string `json:"schema"`
	FixtureID              string `json:"fixture_id"`
	InputSchema            string `json:"input_schema"`
	RegistryArtifactDigest string `json:"registry_artifact_digest"`
	DecisionAlgebra        string `json:"decision_algebra"`
}

type toolchainArtifact struct {
	Schema           string `json:"schema"`
	FixtureID        string `json:"fixture_id"`
	InputSchema      string `json:"input_schema"`
	ToolchainKind    string `json:"toolchain_kind"`
	GoFormat         string `json:"go_format"`
	JSONMode         string `json:"json_mode"`
	SetNormalization string `json:"set_normalization"`
}

type artifactContracts struct {
	authority authorityArtifact
	policy    policyArtifact
	registry  registryArtifact
	oracle    oracleArtifact
	toolchain toolchainArtifact
}

// verifyArtifactBindings parses the immutable contracts, hashes their exact
// bytes, and validates the semantic fields that overlap the evaluator input.
// The registry digest additionally covers every canonical PressureRecord.
func verifyArtifactBindings(input Input) bool {
	contracts, ok := readArtifactContracts()
	if !ok || !contracts.matchInput(input) {
		return false
	}
	return input.AuthoritySnapshotDigest == artifactDigest(authoritySnapshotArtifactBytes) &&
		input.PolicyDigest == artifactDigest(policyArtifactBytes) &&
		input.RegistryDigest == registryBindingDigest(input.PressureRecords, contracts.registry) &&
		input.OracleDigest == artifactDigest(oracleContractArtifactBytes) &&
		input.ToolchainOptionsDigest == artifactDigest(toolchainOptionsArtifactBytes)
}

func readArtifactContracts() (artifactContracts, bool) {
	var contracts artifactContracts
	if decodeStrictJSON(authoritySnapshotArtifactBytes, &contracts.authority) != nil ||
		decodeStrictJSON(policyArtifactBytes, &contracts.policy) != nil ||
		decodeStrictJSON(pressureRegistryArtifactBytes, &contracts.registry) != nil ||
		decodeStrictJSON(oracleContractArtifactBytes, &contracts.oracle) != nil ||
		decodeStrictJSON(toolchainOptionsArtifactBytes, &contracts.toolchain) != nil {
		return artifactContracts{}, false
	}
	if !contracts.valid() {
		return artifactContracts{}, false
	}
	return contracts, true
}

func (contracts artifactContracts) valid() bool {
	if contracts.authority.Schema != "gooo/pressure-independence/fixture-authority-snapshot/v1" ||
		contracts.authority.FixtureID != fixtureID || contracts.authority.InputSchema != SchemaV1 ||
		contracts.authority.Authority != "SYNTHETIC_FIXTURE" || contracts.authority.SourceID == "" ||
		contracts.authority.SemanticID == "" {
		return false
	}
	if contracts.policy.Schema != "gooo/pressure-independence/fixture-policy/v1" ||
		contracts.policy.FixtureID != fixtureID || contracts.policy.InputSchema != SchemaV1 ||
		contracts.policy.PolicyKind != "RESEARCH_ONLY" || contracts.policy.RequestedK == 0 ||
		contracts.policy.MinimumIndependent == 0 {
		return false
	}
	if contracts.registry.Schema != "gooo/pressure-independence/fixture-pressure-registry/v1" ||
		contracts.registry.FixtureID != fixtureID || contracts.registry.InputSchema != SchemaV1 ||
		!canonicalRegistryMatchesArtifact(contracts.registry) {
		return false
	}
	if contracts.oracle.Schema != "gooo/pressure-independence/fixture-oracle-contract/v1" ||
		contracts.oracle.FixtureID != fixtureID || contracts.oracle.InputSchema != SchemaV1 ||
		contracts.oracle.RegistryArtifactDigest != artifactDigest(pressureRegistryArtifactBytes) ||
		contracts.oracle.DecisionAlgebra != "PASS|FAIL_CLOSED|UNKNOWN" {
		return false
	}
	return contracts.toolchain.Schema == "gooo/pressure-independence/fixture-toolchain-options/v1" &&
		contracts.toolchain.FixtureID == fixtureID && contracts.toolchain.InputSchema == SchemaV1 &&
		contracts.toolchain.ToolchainKind == "SYNTHETIC_FIXTURE" && contracts.toolchain.GoFormat == "gofmt" &&
		contracts.toolchain.JSONMode == "strict" && contracts.toolchain.SetNormalization == "sorted"
}

func (contracts artifactContracts) matchInput(input Input) bool {
	return input.Schema == contracts.authority.InputSchema && input.Schema == contracts.policy.InputSchema &&
		input.Schema == contracts.registry.InputSchema && input.Schema == contracts.oracle.InputSchema &&
		input.Schema == contracts.toolchain.InputSchema && strings.HasPrefix(input.FixtureID, "fixture-") &&
		input.RequestedK == contracts.policy.RequestedK &&
		input.MinimumIndependent == contracts.policy.MinimumIndependent
}

func canonicalRegistryMatchesArtifact(registry registryArtifact) bool {
	data, err := canonicalRegistryBytes(registry.Records, registry)
	return err == nil && string(data) == string(pressureRegistryArtifactBytes)
}

func registryBindingDigest(records []PressureRecord, contract registryArtifact) string {
	data, _ := canonicalRegistryBytes(records, contract)
	return digestBytes(data)
}

func canonicalRegistryBytes(records []PressureRecord, contract registryArtifact) ([]byte, error) {
	records = append([]PressureRecord(nil), records...)
	sort.Slice(records, func(left, right int) bool { return pressureKey(records[left]) < pressureKey(records[right]) })
	data, err := json.Marshal(registryArtifact{
		Schema: contract.Schema, FixtureID: contract.FixtureID, InputSchema: contract.InputSchema, Records: records,
	})
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func artifactDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
