package publicpartialreuse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const AssessmentActivityID = "gooo://partial-reuse/assess/v1"
const CaseReportSchema = "gooo/public-partial-test-reuse-report/v2"

// Observation follows Policy.Partitions order. These facts describe a read;
// evaluating them neither reads a receipt nor authorizes a test omission.
type Observation struct {
	ClosureKnown      bool     `json:"closure_known"`
	ScopeBounded      bool     `json:"scope_bounded"`
	MissingEdge       bool     `json:"missing_edge"`
	HiddenChangedEdge bool     `json:"hidden_changed_edge"`
	Impacted          []bool   `json:"impacted"`
	ReceiptStates     []string `json:"receipt_states"`
}

type Assessment struct {
	Decision          string   `json:"decision"`
	Reason            string   `json:"reason"`
	Stage             string   `json:"stage"`
	Step              string   `json:"step"`
	UnknownClass      string   `json:"unknown_class"`
	NextOperation     string   `json:"next_operation"`
	BlockedBy         []string `json:"blocked_by"`
	TestUnitsRequired int      `json:"test_units_required"`
	TestUnitsReusable int      `json:"test_units_reusable"`
}

func (policy Policy) Assess(ctx context.Context, input Observation) (bodycodegen.FixedValue, Assessment, error) {
	if len(policy.decisionSource) == 0 || cache.HashBytes(policy.decisionSource).String() != policy.DecisionSourceDigest {
		return bodycodegen.FixedValue{}, Assessment{}, errors.New("partial reuse decision source identity differs")
	}
	return assessSource(ctx, policy.decisionSource, input)
}

func assessSource(ctx context.Context, source []byte, input Observation) (bodycodegen.FixedValue, Assessment, error) {
	var assessment Assessment
	tuple, err := json.Marshal([]Observation{input})
	if err != nil {
		return bodycodegen.FixedValue{}, assessment, err
	}
	value, err := bodycodegen.EvaluateFixedValue(ctx, "internal/meta/publicpartialreuse/policy.gooo", source, "Assess", tuple)
	if err != nil {
		return value, assessment, fmt.Errorf("evaluate Gooo partial reuse policy: %w", err)
	}
	if value.ActivityID != AssessmentActivityID || value.SourceSHA256 != "sha256:"+cache.HashBytes(source).String() {
		return value, assessment, errors.New("partial reuse assessment origin differs")
	}
	// The in-process typed evaluator constructs this record. Saved records are
	// compared to a fresh source evaluation by VerifyAssessment below.
	if err := json.Unmarshal(value.Value, &assessment); err != nil {
		return value, assessment, err
	}
	return value, assessment, nil
}

func (assessment Assessment) unknown() *UnknownState {
	if assessment.Decision != DecisionUnknown {
		return nil
	}
	return &UnknownState{Stage: assessment.Stage, Step: assessment.Step, Reason: assessment.Reason,
		UnknownClass: assessment.UnknownClass, NextOperation: assessment.NextOperation,
		BlockedBy: append([]string(nil), assessment.BlockedBy...)}
}

// VerifyAssessment re-evaluates the retained observation using the current
// source. It checks report projection; it does not reobserve files or tool inputs.
func VerifyAssessment(report CaseReport) error {
	if report.Schema != CaseReportSchema || report.PolicyEvaluation == nil {
		return errors.New("partial reuse report has no source-owned assessment")
	}
	value, assessment, err := assessSource(context.Background(), decisionSource, report.PolicyInputs)
	if err != nil {
		return err
	}
	actual, _ := json.Marshal(report.PolicyEvaluation)
	expected, _ := json.Marshal(value)
	if !bytes.Equal(actual, expected) {
		return errors.New("partial reuse retained value or origin differs from Gooo source")
	}
	actualUnknown, _ := json.Marshal(report.Unknown)
	expectedUnknown, _ := json.Marshal(assessment.unknown())
	if report.Decision != assessment.Decision || report.Reason != assessment.Reason ||
		!bytes.Equal(actualUnknown, expectedUnknown) || report.After.TestUnitsExecuted != assessment.TestUnitsRequired ||
		report.After.TestUnitsReused != assessment.TestUnitsReusable {
		return errors.New("partial reuse report differs from its source assessment")
	}
	return nil
}
