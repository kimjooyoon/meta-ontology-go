// Package metarecognition is a research-only, read-only comparison fixture.
// It measures the existing registry-bound selective-CI subjects against an
// explicit typed-config baseline; it is not a CI or product implementation.
package metarecognition

import (
	"encoding/json"
	"sort"
)

const SchemaVersion = "gooo/selective-ci-metarecognition/v1"

type Subject string

const (
	SubjectBinding   Subject = "SEMANTIC_BINDING"
	SubjectGraph     Subject = "IMPACT_GRAPH"
	SubjectSoundness Subject = "FULL_SOUNDNESS"
	SubjectPath      Subject = "PATH_CLOSURE"
	SubjectResource  Subject = "RESOURCE_ENVELOPE"
)

type State string

const (
	ClosedSound              State = "CLOSED/SOUND"
	FailClosedUnsound        State = "FAIL_CLOSED/UNSOUND"
	UnknownFullSuiteRequired State = "UNKNOWN/FULL_SUITE_REQUIRED"
)

type Reason string

const (
	ReasonExactBinding      Reason = "EXACT_REGISTERED_BINDING"
	ReasonRenameBinding     Reason = "STABLE_ID_RENAME_EXACT_BINDING"
	ReasonBlobWithoutID     Reason = "FILE_BLOB_WITHOUT_STABLE_ID_BINDING"
	ReasonSourceMapRegistry Reason = "SOURCE_MAP_OR_REGISTRY_MISSING"
	ReasonUnknownGraph      Reason = "UNKNOWN_GRAPH_NODE_OR_MISSED_OBLIGATION"
	ReasonGlobalGuard       Reason = "GLOBAL_GUARD_OMITTED"
	ReasonSelectedDrift     Reason = "SELECTED_FULL_STATUS_OR_OUTPUT_DIGEST_DRIFT"
	ReasonOmittedFailure    Reason = "OMITTED_FULL_SUITE_FAILURE"
	ReasonNonAuthoritative  Reason = "NON_AUTHORITATIVE_OMITTED_OBLIGATION"
	ReasonDuplicateReceipt  Reason = "DUPLICATE_OR_CONFLICTING_PATH_RECEIPT"
	ReasonInvalidResource   Reason = "INVALID_OR_OVERFLOW_RESOURCE_RECEIPT"
	ReasonExternalMissing   Reason = "EXTERNAL_AUTHENTICITY_PROVIDER_PHASE_OBSERVER_MISSING"
)

type Finding string

const (
	NoUniqueBenefit Finding = "NO_UNIQUE_BENEFIT"
	UniqueBenefit   Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
)

type Status string

const (
	Pass Status = "PASS"
	Fail Status = "FAIL"
)

type Authority string

const (
	Authoritative Authority = "AUTHORITATIVE"
	Candidate     Authority = "CANDIDATE"
	Derived       Authority = "DERIVED"
)

// BaselineConfig contains only explicit assertions and stable IDs. It has no
// display-name, prose, inference, timing, or external command surface.
type BaselineConfig struct {
	Subject          Subject
	StableID         string
	BoundID          string
	ExpectedFile     string
	ObservedFile     string
	ExpectedBlob     string
	ObservedBlob     string
	RegistryPresent  bool
	SourceMapPresent bool
	Ambiguous        bool

	UnknownIDs []string
	MissedIDs  []string
	Commands   []CommandAssertion
	Obligation ObligationAssertion
	Path       PathAssertion
	Resource   ResourceAssertion
	External   ExternalAssertion

	FullCommands     int
	SelectedCommands int
	ProvRecords      int
	ProvPaths        int
}

type CommandAssertion struct {
	ID             string `json:"id"`
	FullStatus     Status `json:"full_status"`
	SelectedStatus Status `json:"selected_status"`
	FullDigest     string `json:"full_digest"`
	SelectedDigest string `json:"selected_digest"`
	Selected       bool   `json:"selected"`
	GlobalGuard    bool   `json:"global_guard"`
	Impacted       bool   `json:"impacted"`
	FullFailure    bool   `json:"full_failure"`
}

type ObligationAssertion struct {
	ID        string    `json:"id"`
	Authority Authority `json:"authority"`
	Impacted  bool      `json:"impacted"`
	Selected  bool      `json:"selected"`
}

type PathAssertion struct {
	IDs       []string `json:"ids"`
	Duplicate bool     `json:"duplicate"`
	Conflict  bool     `json:"conflict"`
}

type ResourceAssertion struct {
	Valid    bool `json:"valid"`
	Overflow bool `json:"overflow"`
}

type ExternalAssertion struct {
	Authenticity bool `json:"authenticity"`
	Provider     bool `json:"provider"`
	Phase        bool `json:"phase"`
	Observer     bool `json:"observer"`
}

type Expected struct {
	State        State    `json:"state"`
	Reason       Reason   `json:"reason"`
	LocalizedIDs []string `json:"localized_ids"`
}

type Case struct {
	ID       string         `json:"id"`
	Expected Expected       `json:"expected"`
	Baseline BaselineConfig `json:"baseline"`
}

type Work struct {
	// Units is the deterministic command work count; provenance dimensions are
	// reported separately and are never folded into a weighted scalar.
	Units       int `json:"work_units"`
	Selected    int `json:"selected_commands"`
	Full        int `json:"full_commands"`
	ProvRecords int `json:"prov_records"`
	ProvPaths   int `json:"prov_paths"`
}

type Outcome struct {
	State        State    `json:"state"`
	Reason       Reason   `json:"reason"`
	LocalizedIDs []string `json:"localized_ids"`
	Work         Work     `json:"work"`
}

type ComparisonCase struct {
	ID                    string   `json:"id"`
	Expected              Expected `json:"expected"`
	Gooo                  Outcome  `json:"gooo"`
	Baseline              Outcome  `json:"baseline"`
	ExactOutcomeVector    bool     `json:"exact_outcome_vector"`
	ExactReasonVector     bool     `json:"exact_reason_localization_vector"`
	GoooFalsePass         bool     `json:"gooo_false_pass"`
	GoooFalseNegative     bool     `json:"gooo_false_negative"`
	BaselineFalsePass     bool     `json:"baseline_false_pass"`
	BaselineFalseNegative bool     `json:"baseline_false_negative"`
}

type Ratio struct {
	Selected int  `json:"selected"`
	Full     int  `json:"full"`
	Known    bool `json:"known"`
}

type Summary struct {
	CaseCount                     int   `json:"case_count"`
	ExactOutcomeVector            bool  `json:"exact_outcome_vector"`
	ExactReasonLocalizationVector bool  `json:"exact_reason_localization_vector"`
	GoooFalsePasses               int   `json:"gooo_false_passes"`
	GoooFalseNegatives            int   `json:"gooo_false_negatives"`
	BaselineFalsePasses           int   `json:"baseline_false_passes"`
	BaselineFalseNegatives        int   `json:"baseline_false_negatives"`
	GoooWorkUnits                 int   `json:"gooo_work_units"`
	BaselineWorkUnits             int   `json:"baseline_work_units"`
	GoooRatio                     Ratio `json:"gooo_selected_full_ratio"`
	BaselineRatio                 Ratio `json:"baseline_selected_full_ratio"`
	GoooProvRecords               int   `json:"gooo_prov_records"`
	BaselineProvRecords           int   `json:"baseline_prov_records"`
	GoooProvPaths                 int   `json:"gooo_prov_paths"`
	BaselineProvPaths             int   `json:"baseline_prov_paths"`
}

type Manifest struct {
	Schema  string           `json:"schema"`
	Finding Finding          `json:"finding"`
	Cases   []ComparisonCase `json:"cases"`
	Summary Summary          `json:"summary"`
}

func (s State) Valid() bool {
	return s == ClosedSound || s == FailClosedUnsound || s == UnknownFullSuiteRequired
}

func (r Reason) Valid() bool {
	for _, value := range []Reason{ReasonExactBinding, ReasonRenameBinding, ReasonBlobWithoutID,
		ReasonSourceMapRegistry, ReasonUnknownGraph, ReasonGlobalGuard, ReasonSelectedDrift,
		ReasonOmittedFailure, ReasonNonAuthoritative, ReasonDuplicateReceipt, ReasonInvalidResource,
		ReasonExternalMissing} {
		if r == value {
			return true
		}
	}
	return false
}

func (c Case) normalized() Case {
	c.Expected.LocalizedIDs = sorted(c.Expected.LocalizedIDs)
	c.Baseline.UnknownIDs = sorted(c.Baseline.UnknownIDs)
	c.Baseline.MissedIDs = sorted(c.Baseline.MissedIDs)
	c.Baseline.Path.IDs = sorted(c.Baseline.Path.IDs)
	c.Baseline.Commands = append([]CommandAssertion(nil), c.Baseline.Commands...)
	sort.Slice(c.Baseline.Commands, func(i, j int) bool { return c.Baseline.Commands[i].ID < c.Baseline.Commands[j].ID })
	return c
}

func (m Manifest) CanonicalJSON() ([]byte, error) {
	cases := append([]ComparisonCase(nil), m.Cases...)
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	copyManifest := m
	copyManifest.Cases = cases
	return json.Marshal(copyManifest)
}

func sorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
