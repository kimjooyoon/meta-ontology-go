package workgraph

import "fmt"

const (
	ContractSchema = "gooo/workgraph-project/v1"
	ReportSchema   = "gooo/workgraph-evidence/v1"
	ProjectName    = "gooo-workgraph"
	SourcePath     = "examples/workgraph/main.gooo"
	ClaimID        = "workgraph://claim/reproducible-project-observation"
	ClaimEntity    = "ProjectObservationClaim"
	GateCount      = 7
)

func CanonicalGates() []GateSpec {
	return []GateSpec{
		{"SOURCE_AUTHORITY", "ObserveProject", "SOURCE", "DECLARE_AUTHORITY", "source", "FOUNDATION"},
		{"SYNTAX_ACCEPTED", "CheckProject", "COMPILER", "CHECK_SOURCE", "syntax", "FOUNDATION"},
		{"META_BOUND", "BindProjectEvidence", "META", "BIND_ACTIVITIES", "binding", "COHERENCE"},
		{"DETERMINISTIC_REPLAY", "ReplayProjectGeneration", "GENERATOR", "REPLAY_GENERATION", "replay", "REGRESSION"},
		{"ARTIFACT_GENERATED", "GenerateProjectArtifact", "GENERATOR", "EMIT_ARTIFACT", "artifact", "COHERENCE"},
		{"RESOURCE_OBSERVED", "ObserveProjectResources", "RUNTIME", "SAMPLE_RESOURCES", "resource", "REGRESSION"},
		{"USER_ROUNDTRIP", "CloseProjectClaim", "USER", "CLOSE_CLAIM", "roundtrip", "COHERENCE"},
	}
}

func ValidateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.Project != ProjectName || contract.Source != SourcePath {
		return fmt.Errorf("workgraph contract identity is not canonical")
	}
	wantClaim := ClaimSpec{ID: ClaimID, Entity: ClaimEntity, Authority: SourcePath}
	if contract.Claim != wantClaim {
		return fmt.Errorf("workgraph claim authority is not canonical")
	}
	want := CanonicalGates()
	if len(contract.Gates) != GateCount {
		return fmt.Errorf("workgraph gate denominator = %d, want %d", len(contract.Gates), GateCount)
	}
	for index := range want {
		if contract.Gates[index] != want[index] {
			return fmt.Errorf("workgraph gate %d is not canonical", index)
		}
	}
	return nil
}
