package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

type Summary struct {
	CasesTotal                int `json:"cases_total"`
	CasesPassed               int `json:"cases_passed"`
	TextualChanges            int `json:"textual_changes"`
	StructuralObservations    int `json:"structural_observations"`
	ClaimTransitionCases      int `json:"claim_transition_cases"`
	AdjudicatedCases          int `json:"adjudicated_cases"`
	SemanticPreserved         int `json:"semantic_preserved"`
	SemanticChanged           int `json:"semantic_changed"`
	Indeterminate             int `json:"indeterminate"`
	UnknownPaths              int `json:"unknown_paths"`
	ChangedPathOrContentCount int `json:"changed_path_or_content_count"`
	ClaimsWithExplainedStatus int `json:"claims_with_explained_status"`
	TotalClaims               int `json:"total_claims"`
	ClaimStatusCoverageBPS    int `json:"claim_status_coverage_bps"`
	DistinctPropositions      int `json:"distinct_propositions"`
	AddedClaims               int `json:"added_claims"`
	RemovedClaims             int `json:"removed_claims"`
	ChangedClaims             int `json:"changed_claims"`
	OpenClaims                int `json:"open_claims"`
	DischargedClaims          int `json:"discharged_claims"`
	RefutedClaims             int `json:"refuted_claims"`
	TransitionChains          int `json:"transition_chains"`
	AmbiguousCases            int `json:"ambiguous_cases"`
	ModeledSemanticComponents int `json:"modeled_semantic_components"`
	TotalSemanticComponents   int `json:"total_semantic_components"`
}

type CaseResult struct {
	Definition producer.CaseDefinition `json:"definition"`
	Passed     bool                    `json:"passed"`
	Report     Report                  `json:"report"`
}

type Suite struct {
	Schema                                     string       `json:"schema"`
	SubjectSHA                                 string       `json:"subject_sha"`
	ObservedCheckoutSHA                        string       `json:"observed_checkout_sha"`
	DenominatorID                              string       `json:"denominator_id"`
	DenominatorDigest                          string       `json:"denominator_digest"`
	Decision                                   string       `json:"decision"`
	Resolution                                 string       `json:"resolution"`
	Reason                                     string       `json:"reason,omitempty"`
	ContractReproduction                       string       `json:"contract_reproduction"`
	SubjectSemanticEquivalence                 string       `json:"subject_semantic_equivalence"`
	SourcePaths                                []string     `json:"source_paths"`
	OutputPath                                 string       `json:"output_path"`
	Cases                                      []CaseResult `json:"cases"`
	Summary                                    Summary      `json:"summary"`
	CaseContractCoverageBPS                    int          `json:"case_contract_coverage_bps"`
	DeclaredProjectionComponentKindCoverageBPS int          `json:"declared_projection_component_kind_coverage_bps"`
	SemanticEquivalenceClaim                   string       `json:"semantic_equivalence_claim"`
	MetaSourcePath                             string       `json:"meta_source_path"`
	MetaContractDigest                         string       `json:"meta_contract_digest"`
	DenominatorVersion                         string       `json:"denominator_version"`
	ModeledSemanticComponents                  int          `json:"modeled_semantic_components"`
	TotalSemanticComponents                    int          `json:"total_semantic_components"`
	SuiteDigest                                string       `json:"suite_digest"`
}
