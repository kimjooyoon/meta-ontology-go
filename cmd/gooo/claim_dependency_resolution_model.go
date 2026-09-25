package main

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
