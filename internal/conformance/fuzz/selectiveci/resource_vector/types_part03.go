package resourcevector

import (
	"sort"
)

type Output struct {
	Schema                string   `json:"schema"`
	FixtureID             string   `json:"fixture_id"`
	InputDigest           string   `json:"input_digest"`
	Selected              *Vector  `json:"selected,omitempty"`
	Full                  *Vector  `json:"full,omitempty"`
	Decision              Decision `json:"decision"`
	Reason                Reason   `json:"reason"`
	LimitFailures         []string `json:"limit_failures"`
	FullSuiteRequired     bool     `json:"full_suite_required"`
	ProofValid            bool     `json:"proof_valid"`
	CanonicalOutputDigest string   `json:"canonical_output_digest"`
	ReplayDigest          string   `json:"replay_digest"`
}

// PromotionAuthorized is deliberately not a field in the output wire shape.
// This research oracle cannot authorize promotion.
func (Output) PromotionAuthorized() bool { return false }

type PartialVector struct {
	CPUCoreNS           *uint64 `json:"cpu_core_ns"`
	MemoryBytes         *uint64 `json:"memory_bytes"`
	PeakMemoryBytes     *uint64 `json:"peak_memory_bytes"`
	WorkUnits           *uint64 `json:"work_units"`
	AffectedStableIDs   *uint64 `json:"affected_stable_ids"`
	ApplicablePressures *uint64 `json:"applicable_pressures"`
	IndependentGroups   *uint64 `json:"independent_groups"`
	UniquePROVRecords   *uint64 `json:"unique_prov_records"`
	FinitePROVPaths     *uint64 `json:"finite_prov_paths"`
	ClosureNumerator    *uint64 `json:"closure_numerator"`
	ClosureDenominator  *uint64 `json:"closure_denominator"`
}
type BaselineResult struct {
	Name      string         `json:"name"`
	Decision  Decision       `json:"decision"`
	Reason    Reason         `json:"reason"`
	FullSuite bool           `json:"full_suite"`
	Vector    *PartialVector `json:"vector,omitempty"`
}
type Finding string

const (
	NoUniqueBenefit             Finding = "NO_UNIQUE_BENEFIT"
	UniqueBenefitNotEstablished Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
)

type Comparison struct {
	Oracle               Output         `json:"oracle"`
	TypedConfigFullSuite BaselineResult `json:"typed_config_full_suite"`
	PlainDAGRetry        BaselineResult `json:"plain_dag_retry"`
	Finding              Finding        `json:"finding"`
}

func U64(value uint64) *uint64 { return &value }
func Bool(value bool) *bool    { return &value }
func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
