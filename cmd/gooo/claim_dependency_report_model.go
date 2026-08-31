package main

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
