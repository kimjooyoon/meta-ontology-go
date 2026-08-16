// Package resourcevector contains a strict, standalone replay oracle for
// resource and provenance vectors. It intentionally does not import a
// production selector or copy a producer's aggregate.
package resourcevector

import "sort"

const (
	SchemaV1       = "gooo/selective-ci-resource-vector/v1"
	CorpusSchemaV1 = "gooo/selective-ci-resource-vector-corpus/v1"
)

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionFailClosed Decision = "FAIL_CLOSED"
)

const (
	Pass       = DecisionPass
	Unknown    = DecisionUnknown
	FailClosed = DecisionFailClosed
)

type Reason string

const (
	ReasonNone                  Reason = "NONE"
	ReasonMissingInput          Reason = "MISSING_INPUT"
	ReasonMissingResource       Reason = "MISSING_RESOURCE"
	ReasonMissingPROV           Reason = "MISSING_PROV"
	ReasonInvalidPath           Reason = "INVALID_PATH"
	ReasonDuplicateID           Reason = "DUPLICATE_ID"
	ReasonDuplicateRecord       Reason = "DUPLICATE_PROV_RECORD"
	ReasonDuplicatePath         Reason = "DUPLICATE_PATH"
	ReasonDanglingID            Reason = "DANGLING_ID"
	ReasonSelectionInvalid      Reason = "INVALID_SELECTION"
	ReasonInvalidPressure       Reason = "INVALID_PRESSURE"
	ReasonOverflow              Reason = "COUNT_OVERFLOW"
	ReasonCeilingExceeded       Reason = "CEILING_EXCEEDED"
	ReasonClosureInvalid        Reason = "INVALID_PROV_CLOSURE"
	ReasonRootRelocationInvalid Reason = "INVALID_ROOT_RELOCATION"
)

// PressureRecord is part of a canonical command record. Applicable is a
// pointer so an omitted field cannot be mistaken for false.
type PressureRecord struct {
	ID                  string `json:"id"`
	IndependenceGroupID string `json:"independence_group_id"`
	Applicable          *bool  `json:"applicable"`
}

// CommandRecord is the canonical command-side record. Resource dimensions
// are presence-aware: nil means the producer did not provide the field.
type CommandRecord struct {
	ID              string           `json:"id"`
	Path            string           `json:"path"`
	CPUCoreNS       *uint64          `json:"cpu_core_ns"`
	MemoryBytes     *uint64          `json:"memory_bytes"`
	PeakMemoryBytes *uint64          `json:"peak_memory_bytes"`
	WorkUnits       *uint64          `json:"work_units"`
	Pressures       []PressureRecord `json:"pressures"`
}

// PathRecord is the canonical provenance-side record. Record IDs are kept as
// an ordered list in the input, but the oracle counts them only after checking
// global uniqueness. No duplicate is silently collapsed.
type PathRecord struct {
	ID                 string   `json:"id"`
	Path               string   `json:"path"`
	CommandID          string   `json:"command_id"`
	RecordIDs          []string `json:"record_ids"`
	Finite             *bool    `json:"finite"`
	ClosureNumerator   *uint64  `json:"closure_numerator"`
	ClosureDenominator *uint64  `json:"closure_denominator"`
}

// CeilingSet contains an explicit ceiling for every replay dimension.
// Pointers distinguish an omitted ceiling from an intentional numeric value.
type CeilingSet struct {
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

type ResourceCeilings struct {
	Selected CeilingSet `json:"selected"`
	Full     CeilingSet `json:"full"`
}

// Input contains only producer-side canonical records. Expected labels are
// held by CorpusCase and are intentionally not part of this value.
type Input struct {
	Schema             string           `json:"schema"`
	FixtureID          string           `json:"fixture_id"`
	Root               string           `json:"root"`
	Commands           []CommandRecord  `json:"commands"`
	Paths              []PathRecord     `json:"paths"`
	SelectedCommandIDs []string         `json:"selected_command_ids"`
	FullCommandIDs     []string         `json:"full_command_ids"`
	Ceilings           ResourceCeilings `json:"ceilings"`
}

// Vector is a fully recomputed aggregate. A vector is present in Output only
// when all fields needed to compute it are present and finite.
type Vector struct {
	CPUCoreNS           uint64 `json:"cpu_core_ns"`
	MemoryBytes         uint64 `json:"memory_bytes"`
	PeakMemoryBytes     uint64 `json:"peak_memory_bytes"`
	WorkUnits           uint64 `json:"work_units"`
	AffectedStableIDs   uint64 `json:"affected_stable_ids"`
	ApplicablePressures uint64 `json:"applicable_pressures"`
	IndependentGroups   uint64 `json:"independent_groups"`
	UniquePROVRecords   uint64 `json:"unique_prov_records"`
	FinitePROVPaths     uint64 `json:"finite_prov_paths"`
	ClosureNumerator    uint64 `json:"closure_numerator"`
	ClosureDenominator  uint64 `json:"closure_denominator"`
}

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

func Bool(value bool) *bool { return &value }

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
