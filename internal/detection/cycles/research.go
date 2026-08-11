package cycles

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Outcome is the result of evaluating a current Go-hosted fixture.
type Outcome string

const (
	OutcomePass     Outcome = "pass"
	OutcomeFail     Outcome = "fail"
	OutcomeDeferred Outcome = "deferred"
)

// Criteria records explicit review criteria instead of hiding them in test
// code. Deferred criteria describe work that is intentionally not claimed.
type Criteria struct {
	Pass     []string `json:"pass"`
	Fail     []string `json:"fail"`
	Deferred []string `json:"deferred"`
}

// FollowUpContract describes the reusable boundary for a later host or tool.
type FollowUpContract struct {
	CurrentHost HostStage `json:"current_host"`
	FutureHost  HostStage `json:"future_host"`
	Input       string    `json:"input"`
	Output      string    `json:"output"`
	Invariants  []string  `json:"invariants"`
	Deferred    []string  `json:"deferred"`
}

// ResearchFixture is a falsifiable graph hypothesis plus its expected result.
type ResearchFixture struct {
	Name           string           `json:"name"`
	Hypothesis     string           `json:"hypothesis"`
	Counterexample string           `json:"counterexample"`
	Graph          Graph            `json:"graph"`
	Expected       map[Code]int     `json:"expected"`
	Criteria       Criteria         `json:"criteria"`
	FollowUp       FollowUpContract `json:"follow_up"`
}

// Measurement is the deterministic, machine-readable output of a fixture.
type Measurement struct {
	NodeCount       int          `json:"node_count"`
	EdgeCount       int          `json:"edge_count"`
	DiagnosticCount int          `json:"diagnostic_count"`
	CodeCounts      map[Code]int `json:"code_counts"`
	Digest          string       `json:"digest"`
}

// Evaluation retains both the measured values and deferred work for CI or
// later research consumers.
type Evaluation struct {
	Outcome  Outcome     `json:"outcome"`
	Measure  Measurement `json:"measure"`
	Reason   string      `json:"reason"`
	Deferred []string    `json:"deferred"`
}

// LoadResearchFixture decodes a strict JSON fixture contract.
func LoadResearchFixture(data []byte) (ResearchFixture, error) {
	var fixture ResearchFixture
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil {
		return ResearchFixture{}, fmt.Errorf("decode research fixture: %w", err)
	}
	if err := fixture.Validate(); err != nil {
		return ResearchFixture{}, err
	}
	return fixture, nil
}

// LoadFixture is a short alias for LoadResearchFixture.
func LoadFixture(data []byte) (ResearchFixture, error) {
	return LoadResearchFixture(data)
}

// Validate checks that a fixture contains falsifiable and non-aspirational
// metadata before its graph is evaluated.
func (f ResearchFixture) Validate() error {
	if strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.Hypothesis) == "" {
		return fmt.Errorf("fixture name and hypothesis are required")
	}
	for code, count := range f.Expected {
		if count < 0 {
			return fmt.Errorf("expected count for %q is negative", code)
		}
	}
	if err := validateCriteria(f.Criteria); err != nil {
		return err
	}
	return f.FollowUp.Validate()
}

func validateCriteria(criteria Criteria) error {
	if len(criteria.Pass) == 0 || len(criteria.Fail) == 0 || len(criteria.Deferred) == 0 {
		return fmt.Errorf("pass, fail, and deferred criteria are required")
	}
	return nil
}

func (c FollowUpContract) Validate() error {
	if c.CurrentHost != GoHostedStage || c.FutureHost != GoooHostedStage {
		return fmt.Errorf("follow-up host stages must be go-hosted and gooo-hosted")
	}
	if strings.TrimSpace(c.Input) == "" || strings.TrimSpace(c.Output) == "" {
		return fmt.Errorf("follow-up input and output contracts are required")
	}
	if len(c.Invariants) == 0 || len(c.Deferred) == 0 {
		return fmt.Errorf("follow-up invariants and deferred work are required")
	}
	return nil
}

// Measure evaluates the current Go-hosted detector and returns stable counts
// and evidence digest. It does not evaluate the future host.
func Measure(graph Graph) Measurement {
	diagnostics := Detect(graph)
	evidence := NewEvidence(GoHostedStage, "research-fixture", diagnostics)
	counts := make(map[Code]int)
	for _, diagnostic := range diagnostics {
		counts[diagnostic.Code]++
	}
	return Measurement{
		NodeCount: len(graph.Nodes), EdgeCount: len(graph.edges()),
		DiagnosticCount: len(diagnostics), CodeCounts: counts, Digest: evidence.Digest,
	}
}

// Evaluate compares measured diagnostic counts with the fixture expectation.
func (f ResearchFixture) Evaluate() Evaluation {
	measurement := Measure(f.Graph)
	if equalCounts(measurement.CodeCounts, f.Expected) {
		return Evaluation{Outcome: OutcomePass, Measure: measurement, Deferred: f.FollowUp.Deferred}
	}
	return Evaluation{
		Outcome: OutcomeFail, Measure: measurement,
		Reason:   fmt.Sprintf("expected codes %#v, got %#v", f.Expected, measurement.CodeCounts),
		Deferred: f.FollowUp.Deferred,
	}
}

func equalCounts(left, right map[Code]int) bool {
	if len(left) != len(right) {
		return false
	}
	for code, count := range left {
		if right[code] != count {
			return false
		}
	}
	return true
}
