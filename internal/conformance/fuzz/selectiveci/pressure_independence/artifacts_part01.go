package pressureindependence

import (
	_ "embed"
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
