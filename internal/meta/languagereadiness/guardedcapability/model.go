package guardedcapability

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"

type Source struct {
	CurrentHeadSHA        string                  `json:"current_head_sha"`
	WorkflowRunID         int64                   `json:"foundation_workflow_run_id"`
	ArtifactID            int64                   `json:"foundation_artifact_id"`
	ArtifactDigest        string                  `json:"foundation_artifact_digest"`
	ReportFileSHA         string                  `json:"foundation_report_file_sha256"`
	FoundationReport     guardedpromotion.Report `json:"foundation_report"`
	AncestryObserved      bool                    `json:"ancestry_observed"`
	FoundationAncestor    bool                    `json:"foundation_ancestor"`
	GuardTreesObserved    bool                    `json:"guard_trees_observed"`
	FoundationGuardTree   string                  `json:"foundation_guard_tree"`
	CurrentGuardTree      string                  `json:"current_guard_tree"`
	WitnessTreesObserved  bool                    `json:"witness_trees_observed"`
	FoundationWitnessTree string                  `json:"foundation_witness_tree"`
	CurrentWitnessTree    string                  `json:"current_witness_tree"`
	RepositoryWrites      int                     `json:"repository_writes"`
	MutationAuthorized    bool                    `json:"repository_mutation_authorized"`
}

type Summary struct {
	Satisfied             int  `json:"satisfied"`
	Total                 int  `json:"total"`
	NotSatisfied          int  `json:"not_satisfied"`
	Unresolved            int  `json:"unresolved"`
	ReadinessBPS          int  `json:"readiness_bps"`
	FoundationReceipts    int  `json:"foundation_receipts"`
	EquivalentTrees       int  `json:"equivalent_implementation_trees"`
	RepositoryWrites      int  `json:"repository_writes"`
	MutationAuthorized    bool `json:"repository_mutation_authorized"`
}

type Receipt struct {
	Schema       string                       `json:"schema"`
	Decision     string                       `json:"decision"`
	Reason       string                       `json:"reason"`
	Resolution   string                       `json:"resolution"`
	Source       Source                       `json:"source"`
	Summary      Summary                      `json:"summary"`
	Coordinates  []guardedpromotion.Coordinate `json:"coordinates"`
	Indicators   []guardedpromotion.Indicator  `json:"indicators"`
	Proofs       []guardedpromotion.Proof      `json:"proofs"`
	ReportDigest string                       `json:"report_digest"`
}
