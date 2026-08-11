// Package roundtripcontract defines the evidence boundary for round-trip
// detection. It is intentionally independent from syntax, semantic IR,
// bidirectional lenses, generators, analyzers, and CI transports.
//
// A detector adapter supplies ArtifactRefs, Findings, and Measurements. The
// contract decides whether a scenario passed, failed, or is deferred; a
// deferred stage is never merge-eligible.
package roundtripcontract

import (
	"fmt"
	"sort"
	"strings"
)

const Version = "gooo/roundtrip-evidence/v1"

// Stage identifies an authoritative or derived projection boundary.
type Stage string

const (
	StageDSL      Stage = "dsl"
	StageIR       Stage = "ir"
	StageGo       Stage = "go"
	StageLiftedIR Stage = "lifted-ir"
	StageEvidence Stage = "evidence"
)

// Outcome is the only status vocabulary accepted by the gate contract.
type Outcome string

const (
	OutcomePass     Outcome = "pass"
	OutcomeFail     Outcome = "fail"
	OutcomeDeferred Outcome = "deferred"
)

// Hypothesis is a falsifiable research claim and its explicit counterexample.
type Hypothesis struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Falsifier string `json:"falsifier"`
}

// ArtifactRef connects evidence to a source, projection, or cache artifact.
// Digest is intentionally opaque so adapters may use SHA-256 or a repository
// content-addressed digest without changing this contract.
type ArtifactRef struct {
	Stage  Stage  `json:"stage"`
	URI    string `json:"uri"`
	Format string `json:"format"`
	Digest string `json:"digest"`
}

// Finding is one deterministic violation emitted by a detector adapter.
type Finding struct {
	Rule     string `json:"rule"`
	Path     string `json:"path"`
	Identity string `json:"identity,omitempty"`
	Detail   string `json:"detail"`
}

// Measurement records the smallest useful reproducibility envelope.
type Measurement struct {
	Nodes         int64 `json:"nodes"`
	Facts         int64 `json:"facts"`
	Regions       int64 `json:"regions"`
	SourceBytes   int64 `json:"source_bytes"`
	Checks        int64 `json:"checks"`
	DurationNanos int64 `json:"duration_nanos"`
}

// Evidence is the detector output consumed by CI, provenance, and cache
// layers. It contains no assertion that an unimplemented stage succeeded.
type Evidence struct {
	Version        string        `json:"version"`
	CaseID         string        `json:"case_id"`
	HypothesisID   string        `json:"hypothesis_id"`
	Outcome        Outcome       `json:"outcome"`
	Artifacts      []ArtifactRef `json:"artifacts"`
	Findings       []Finding     `json:"findings"`
	Measurement    Measurement   `json:"measurement"`
	DeferredReason string        `json:"deferred_reason,omitempty"`
}

// Scenario binds a falsifiable hypothesis to its expected research outcome.
type Scenario struct {
	CaseID     string     `json:"case_id"`
	Hypothesis Hypothesis `json:"hypothesis"`
	Mutation   string     `json:"mutation"`
	Expected   Outcome    `json:"expected"`
	Evidence   Evidence   `json:"evidence"`
}

// Assessment separates “the hypothesis was classified correctly” from
// “the change may proceed”. Fail and deferred scenarios can be valid research
// evidence while remaining ineligible for merge.
type Assessment struct {
	CaseID        string  `json:"case_id"`
	Expected      Outcome `json:"expected"`
	Actual        Outcome `json:"actual"`
	Accepted      bool    `json:"accepted"`
	MergeEligible bool    `json:"merge_eligible"`
	Detail        string  `json:"detail"`
}

// Validate checks the contract and that Outcome agrees with evidence content.
func (e Evidence) Validate() error {
	if e.Version != Version {
		return fmt.Errorf("unsupported evidence version %q", e.Version)
	}
	if strings.TrimSpace(e.CaseID) == "" || strings.TrimSpace(e.HypothesisID) == "" {
		return fmt.Errorf("case and hypothesis IDs are required")
	}
	if !validOutcome(e.Outcome) {
		return fmt.Errorf("unknown outcome %q", e.Outcome)
	}
	if err := e.Measurement.validate(); err != nil {
		return err
	}
	if err := validateArtifacts(e.Artifacts); err != nil {
		return err
	}
	if err := validateFindings(e.Findings); err != nil {
		return err
	}
	if e.Outcome != derivedOutcome(e) {
		return fmt.Errorf("outcome %q disagrees with findings/deferred reason", e.Outcome)
	}
	if e.Outcome != OutcomeDeferred && strings.TrimSpace(e.DeferredReason) != "" {
		return fmt.Errorf("deferred reason is only valid for deferred evidence")
	}
	return nil
}

// Validate checks the expected outcome and its evidence identity.
func (s Scenario) Validate() error {
	if strings.TrimSpace(s.CaseID) == "" || strings.TrimSpace(s.Mutation) == "" {
		return fmt.Errorf("scenario case ID and mutation are required")
	}
	if strings.TrimSpace(s.Hypothesis.ID) == "" || strings.TrimSpace(s.Hypothesis.Statement) == "" || strings.TrimSpace(s.Hypothesis.Falsifier) == "" {
		return fmt.Errorf("hypothesis ID, statement, and falsifier are required")
	}
	if !validOutcome(s.Expected) {
		return fmt.Errorf("unknown expected outcome %q", s.Expected)
	}
	if s.Evidence.CaseID != s.CaseID || s.Evidence.HypothesisID != s.Hypothesis.ID {
		return fmt.Errorf("evidence identity does not match scenario")
	}
	return s.Evidence.Validate()
}

// Assess classifies one scenario without treating deferred work as success.
func Assess(s Scenario) (Assessment, error) {
	if err := s.Validate(); err != nil {
		return Assessment{}, err
	}
	actual := s.Evidence.Outcome
	accepted := actual == s.Expected
	detail := "expected outcome observed"
	if !accepted {
		detail = fmt.Sprintf("expected %s, observed %s", s.Expected, actual)
	}
	return Assessment{
		CaseID:        s.CaseID,
		Expected:      s.Expected,
		Actual:        actual,
		Accepted:      accepted,
		MergeEligible: accepted && actual == OutcomePass,
		Detail:        detail,
	}, nil
}

func (m Measurement) validate() error {
	values := []struct {
		name  string
		value int64
	}{
		{"nodes", m.Nodes}, {"facts", m.Facts}, {"regions", m.Regions},
		{"source bytes", m.SourceBytes}, {"checks", m.Checks}, {"duration", m.DurationNanos},
	}
	for _, value := range values {
		if value.value < 0 {
			return fmt.Errorf("measurement %s is negative", value.name)
		}
	}
	return nil
}

func validateArtifacts(artifacts []ArtifactRef) error {
	seen := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Stage == "" || strings.TrimSpace(artifact.URI) == "" || strings.TrimSpace(artifact.Format) == "" || strings.TrimSpace(artifact.Digest) == "" {
			return fmt.Errorf("artifact stage, URI, format, and digest are required")
		}
		key := string(artifact.Stage) + "\x00" + artifact.URI
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate artifact %q", artifact.URI)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateFindings(findings []Finding) error {
	for _, finding := range findings {
		if strings.TrimSpace(finding.Rule) == "" || strings.TrimSpace(finding.Path) == "" || strings.TrimSpace(finding.Detail) == "" {
			return fmt.Errorf("finding rule, path, and detail are required")
		}
	}
	return nil
}

func derivedOutcome(e Evidence) Outcome {
	if strings.TrimSpace(e.DeferredReason) != "" {
		return OutcomeDeferred
	}
	if len(e.Findings) > 0 {
		return OutcomeFail
	}
	return OutcomePass
}

func validOutcome(outcome Outcome) bool {
	return outcome == OutcomePass || outcome == OutcomeFail || outcome == OutcomeDeferred
}

func normalizedFindings(findings []Finding) []Finding {
	result := append([]Finding(nil), findings...)
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Rule != right.Rule {
			return left.Rule < right.Rule
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Identity != right.Identity {
			return left.Identity < right.Identity
		}
		return left.Detail < right.Detail
	})
	return result
}
