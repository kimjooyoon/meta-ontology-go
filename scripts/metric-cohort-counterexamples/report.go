package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

type report struct {
	Schema                  string                                `json:"schema"`
	Decision                string                                `json:"decision"`
	Origin                  origin                                `json:"origin"`
	ProducerBinding         string                                `json:"producer_binding"`
	ExecutionOrigin         *executionOrigin                      `json:"execution_origin,omitempty"`
	RegistryOrigin          metricprogram.RegistryOrigin          `json:"registry_origin"`
	DeclaredCohort          []metricstrategy.ConceptOperationSpec `json:"declared_cohort"`
	DeclaredCohortDigest    string                                `json:"declared_cohort_digest"`
	ExpectedCount           int                                   `json:"expected_count"`
	CasesDeclared           int                                   `json:"cases_declared"`
	CasesAttempted          int                                   `json:"cases_attempted"`
	CasesExecuted           int                                   `json:"cases_executed"`
	CasesPassed             int                                   `json:"cases_passed"`
	SetupElapsedNanoseconds int64                                 `json:"setup_elapsed_nanoseconds"`
	ElapsedNanoseconds      int64                                 `json:"elapsed_nanoseconds"`
	ElapsedScope            string                                `json:"elapsed_scope"`
	ReadinessTotal          int                                   `json:"readiness_total"`
	VariantPolicy           string                                `json:"variant_policy"`
	UnrelatedScope          string                                `json:"unrelated_scope"`
	RepositoryWrites        int                                   `json:"repository_writes"`
	PromotionAuthorized     bool                                  `json:"promotion_authorized"`
	SetupError              string                                `json:"setup_error,omitempty"`
	Cases                   []caseResult                          `json:"cases"`
}
