package publictestreuse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// ReuseObservation contains observations made by the caller's file and receipt
// readers. Its values alone never authorize an effect or establish closure.
type ReuseObservation struct {
	Requested    bool   `json:"requested"`
	ReceiptState string `json:"receipt_state"`
	PolicyState  string `json:"policy_state"`
	BindingState string `json:"binding_state"`
	InputProfile string `json:"input_profile"`
}

type OriginalObservation struct {
	BuildPassed     bool `json:"build_passed"`
	TestPassed      bool `json:"test_passed"`
	SemanticMatches bool `json:"semantic_matches"`
}

// Assessment is the value produced by the fixed Gooo activity. The host keeps
// responsibility for observing the inputs and validating any proposed reuse.
type Assessment struct {
	CaseID        string `json:"case_id"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	Step          string `json:"step"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
	Reuse         bool   `json:"reuse"`
}

func (policy Policy) AssessReplay(ctx context.Context, input ReuseObservation) (bodycodegen.FixedValue, Assessment, error) {
	return policy.assess(ctx, "AssessReplay", "gooo://test-reuse/assess-replay/v1", input)
}

func (policy Policy) AssessOriginal(ctx context.Context, input OriginalObservation) (bodycodegen.FixedValue, Assessment, error) {
	return policy.assess(ctx, "AssessOriginal", "gooo://test-reuse/assess-original/v1", input)
}

func (policy Policy) assess(ctx context.Context, activity, activityID string, input any) (bodycodegen.FixedValue, Assessment, error) {
	var assessment Assessment
	if cache.HashBytes(policy.source).String() != policy.DecisionSourceDigest {
		return bodycodegen.FixedValue{}, assessment, errors.New("policy source differs from its retained identity")
	}
	tuple, err := json.Marshal([]any{input})
	if err != nil {
		return bodycodegen.FixedValue{}, assessment, err
	}
	value, err := bodycodegen.EvaluateFixedValue(ctx, policy.filename, policy.source, activity, tuple)
	if err != nil {
		return value, assessment, fmt.Errorf("evaluate Gooo reuse policy: %w", err)
	}
	if value.ActivityID != activityID || value.SourceSHA256 != policy.DecisionSourceDigest || !cache.Digest(value.ProjectionSHA256).Known() {
		return value, assessment, errors.New("Gooo reuse policy evaluation identity differs")
	}
	if err := decodeAssessmentObject(value.Value, &assessment); err != nil {
		return value, assessment, fmt.Errorf("decode Gooo reuse assessment: %w", err)
	}
	if err := assessment.validate(); err != nil {
		return value, assessment, err
	}
	return value, assessment, nil
}

// The fixed policy boundary carries complete scalar records. Keep missing,
// null, duplicate and differently cased fields distinct from explicit values.
func decodeAssessmentObject(data []byte, target any) error {
	if err := completeness.CheckJSON(data, 16<<10); err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	canonical, err := json.Marshal(target)
	if err != nil {
		return err
	}
	var observed, expected map[string]json.RawMessage
	if err := json.Unmarshal(data, &observed); err != nil {
		return err
	}
	if err := json.Unmarshal(canonical, &expected); err != nil {
		return err
	}
	if len(observed) != len(expected) {
		return errors.New("Gooo assessment record has incomplete or unknown fields")
	}
	for key := range expected {
		value, present := observed[key]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("Gooo assessment field %q is missing or null", key)
		}
	}
	return nil
}

func (assessment Assessment) validate() error {
	if assessment.CaseID == "" || len(assessment.CaseID) > 128 || assessment.Reason == "" || len(assessment.Reason) > 256 {
		return errors.New("Gooo reuse assessment has no bounded case and reason")
	}
	switch assessment.Decision {
	case DecisionUnknown:
		if assessment.Step == "" || assessment.NextOperation == "" || assessment.BlockedBy == "" || assessment.Reuse {
			return errors.New("unresolved Gooo assessment lacks its next operation")
		}
	case DecisionClosed, DecisionRefuted:
		if assessment.Step != "" || assessment.NextOperation != "" || assessment.BlockedBy != "" ||
			(assessment.Decision == DecisionRefuted && assessment.Reuse) {
			return errors.New("Gooo reuse assessment contradicts its decision")
		}
	default:
		return errors.New("Gooo reuse assessment has an unknown decision")
	}
	if len(assessment.Step) > 128 || len(assessment.NextOperation) > 256 || len(assessment.BlockedBy) > 256 {
		return errors.New("Gooo reuse assessment diagnostic exceeds its bounds")
	}
	return nil
}
