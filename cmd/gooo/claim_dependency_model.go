package main

const (
	claimDependencySchema      = "gooo/claim-dependency-causality/v1"
	claimDependencyCandidateID = "gooo.primitive.claim-dependency-causality.v1"
	claimDependencyObserved    = "CLAIM_DEPENDENCY_OBSERVED"
	claimDependencyIncomplete  = "INCOMPLETE"
	claimDependencyFailed      = "FAIL_CLOSED"
	claimDependencyRootRole    = "RECOVERABLE_ROOT"
	claimDependencyEdgeRole    = "TYPED_DEPENDENCY"
	claimDependencyFoundation  = "FOUNDATION"
	claimDependencyCoherence   = "COHERENCE"
	claimDependencyRegression  = "REGRESSION"
)

var claimDependencyEdgeKinds = []string{"REQUIRES", "SUPPORTS", "CONTRADICTS", "FAILURE_ENTAILMENT"}

type claimDependencySubject struct {
	SourceFile   string `json:"source_file"`
	SourceDigest string `json:"source_digest"`
	Namespace    string `json:"namespace"`
	Binding      string `json:"binding"`
}

type claimDependencyContract struct {
	Version            string   `json:"version"`
	RootProgram        string   `json:"root_program"`
	EdgeProgramPrefix  string   `json:"edge_program_prefix"`
	EdgeKinds          []string `json:"edge_kinds"`
	GraphSemantics     string   `json:"graph_semantics"`
	CyclePolicy        string   `json:"cycle_policy"`
	MissingInputPolicy string   `json:"missing_input_policy"`
}

type claimDependencyNode struct {
	Ordinal            int    `json:"ordinal"`
	Activity           string `json:"activity"`
	OutputEntity       string `json:"output_entity"`
	Role               string `json:"role"`
	Label              string `json:"label"`
	ProofChoice        string `json:"proof_choice"`
	ValueProgramDigest string `json:"value_program_digest"`
}

type claimDependencyEdge struct {
	Ordinal      int    `json:"ordinal"`
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Label        string `json:"label"`
	FromActivity string `json:"from_activity"`
	ToActivity   string `json:"to_activity"`
	ViaEntity    string `json:"via_entity"`
}
