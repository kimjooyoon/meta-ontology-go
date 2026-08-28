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

type claimDependencyGap struct {
	Activity      string `json:"activity"`
	InputEntity   string `json:"input_entity"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
}

type claimDependencyKindCounts struct {
	Requires          int `json:"requires"`
	Supports          int `json:"supports"`
	Contradicts       int `json:"contradicts"`
	FailureEntailment int `json:"failure_entailment"`
}

type claimDependencySummary struct {
	ActivitiesTotal    int `json:"activities_total"`
	ActivitiesObserved int `json:"activities_observed"`
	RecoverableRoots   int `json:"recoverable_roots"`
	TypedDeclarations  int `json:"typed_declarations"`
	DependencyInputs   int `json:"dependency_inputs"`
	TypedEdges         int `json:"typed_edges"`
	EdgeKindsObserved  int `json:"edge_kinds_observed"`
	UnresolvedInputs   int `json:"unresolved_inputs"`
	CyclicActivities   int `json:"cyclic_activities"`
	RepositoryWrites   int `json:"repository_writes"`
}

type claimDependencyResolution struct {
	State         string   `json:"state"`
	Stage         *string  `json:"stage"`
	Step          *string  `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  *string  `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	ProofChoice   string   `json:"proof_choice"`
	BlockedBy     []string `json:"blocked_by"`
}

type claimDependencyIndicator struct {
	ID          string   `json:"id"`
	Value       int      `json:"value"`
	Target      int      `json:"target"`
	Comparator  string   `json:"comparator"`
	Unit        string   `json:"unit"`
	Class       string   `json:"class"`
	ProofChoice string   `json:"proof_choice"`
	Activities  []string `json:"activities"`
}

type claimDependencyAuthority struct {
	Source                     string `json:"source"`
	SemanticTruthClaimed       bool   `json:"semantic_truth_claimed"`
	StatePropagationAuthorized bool   `json:"state_propagation_authorized"`
	CoreMutationAuthorized     bool   `json:"core_mutation_authorized"`
	AutomaticMergeAllowed      bool   `json:"automatic_merge_allowed"`
	RepositoryWrites           int    `json:"repository_writes"`
}

type claimDependencyReport struct {
	Schema     string                     `json:"schema"`
	Candidate  string                     `json:"candidate_id"`
	Decision   string                     `json:"decision"`
	Subject    claimDependencySubject     `json:"subject"`
	Contract   claimDependencyContract    `json:"contract"`
	Resolution claimDependencyResolution  `json:"resolution"`
	Summary    claimDependencySummary     `json:"summary"`
	KindCounts claimDependencyKindCounts  `json:"kind_counts"`
	Nodes      []claimDependencyNode      `json:"nodes"`
	Edges      []claimDependencyEdge      `json:"edges"`
	Gaps       []claimDependencyGap       `json:"gaps"`
	Indicators []claimDependencyIndicator `json:"indicators"`
	Authority  claimDependencyAuthority   `json:"authority"`
}
