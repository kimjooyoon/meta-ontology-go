package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

type Summary struct {
	CasesTotal             int `json:"cases_total"`
	CasesPassed            int `json:"cases_passed"`
	TextualChanges         int `json:"textual_changes"`
	StructuralObservations int `json:"structural_observations"`
	ClaimTransitionCases   int `json:"claim_transition_cases"`
	AdjudicatedCases       int `json:"adjudicated_cases"`
	SemanticPreserved      int `json:"semantic_preserved"`
	SemanticChanged        int `json:"semantic_changed"`
	Indeterminate          int `json:"indeterminate"`
	UnknownPaths           int `json:"unknown_paths"`
	RepositoryWrites       int `json:"repository_writes"`
}

type CaseResult struct {
	Definition producer.CaseDefinition `json:"definition"`
	Passed     bool                    `json:"passed"`
	Report     Report                  `json:"report"`
}

type Suite struct {
	Schema            string       `json:"schema"`
	SubjectSHA        string       `json:"subject_sha"`
	DenominatorID     string       `json:"denominator_id"`
	DenominatorDigest string       `json:"denominator_digest"`
	Decision          string       `json:"decision"`
	Resolution        string       `json:"resolution"`
	Cases             []CaseResult `json:"cases"`
	Summary           Summary      `json:"summary"`
	CoverageBPS       int          `json:"coverage_bps"`
	SuiteDigest       string       `json:"suite_digest"`
}
