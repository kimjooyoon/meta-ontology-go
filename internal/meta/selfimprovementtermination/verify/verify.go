package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	termination "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination"
)

const ReportSchema = "gooo/self-improvement-termination-judge/v1"

type Report struct {
	Schema             string                        `json:"schema"`
	Independent        bool                          `json:"independent"`
	Decision           string                        `json:"decision"`
	Reason             string                        `json:"reason"`
	ReceiptDigest      string                        `json:"receipt_digest"`
	Satisfied          int                           `json:"satisfied"`
	Total              int                           `json:"total"`
	BasisPoints        int                           `json:"basis_points"`
	ClaimTransitions   []termination.ClaimTransition `json:"claim_transitions"`
	VerificationDigest string                        `json:"verification_digest"`
}

type classification struct {
	decision, reason, finalState string
	period, stateCount           int
	hasCycle, diverging          bool
}

func Verify(input termination.Input, receipt termination.Receipt) (Report, error) {
	if err := validInput(input); err != nil {
		return Report{}, err
	}
	class := classify(input.Trace, input.MaxSteps)
	if err := validReceipt(input, receipt, class); err != nil {
		return Report{}, err
	}
	report := Report{
		Schema: ReportSchema, Independent: true, Decision: class.decision, Reason: class.reason,
		ReceiptDigest: receipt.ReceiptDigest, Satisfied: receipt.Summary.Satisfied,
		Total: receipt.Summary.Total, BasisPoints: receipt.Summary.BasisPoints,
		ClaimTransitions: append([]termination.ClaimTransition(nil), receipt.ClaimTransitions...),
	}
	report.VerificationDigest = digestJSON(struct {
		Decision, Reason, ReceiptDigest string
	}{report.Decision, report.Reason, report.ReceiptDigest})
	return report, nil
}

func validInput(input termination.Input) error {
	if input.Schema != termination.InputSchema || input.Repository == "" || input.Subject != termination.Consumer ||
		input.Producer != termination.Producer || input.Consumer != termination.Consumer ||
		input.MetaOperation != termination.MetaOperation || input.ProofChoice != termination.ProofChoice ||
		input.Stage != termination.TraceStage || input.MaxSteps < 1 || input.MaxSteps > termination.MaxTraceSteps ||
		len(input.Trace) == 0 || len(input.Trace) > input.MaxSteps {
		return fmt.Errorf("independent judge: input identity or budget is not bound")
	}
	for index, observation := range input.Trace {
		if observation.Stage != termination.TraceStage || observation.Step != index+1 ||
			observation.BeforeRank < 0 || observation.AfterRank < 0 ||
			!validDigest(observation.BeforeState) || !validDigest(observation.AfterState) {
			return fmt.Errorf("independent judge: malformed step %d", index+1)
		}
		if index > 0 && input.Trace[index-1].AfterState != observation.BeforeState {
			return fmt.Errorf("independent judge: broken state chain at step %d", index+1)
		}
		changed := observation.BeforeState != observation.AfterState
		if changed != (observation.Decision == "CHANGED") {
			return fmt.Errorf("independent judge: decision mismatch at step %d", index+1)
		}
		if changed && observation.Reason != "METAPROGRAM_STATE_CHANGED" {
			return fmt.Errorf("independent judge: changed reason mismatch at step %d", index+1)
		}
		if !changed && (observation.Decision != "NO_CHANGE" || observation.Reason != "NO_CHANGE_FIXED_POINT_OBSERVED") {
			return fmt.Errorf("independent judge: no-change reason mismatch at step %d", index+1)
		}
	}
	return nil
}

func validReceipt(input termination.Input, receipt termination.Receipt, class classification) error {
	if receipt.Schema != termination.ReceiptSchema || receipt.Metaprogram != termination.Metaprogram ||
		receipt.Status != termination.ReceiptBound || receipt.Resolution != termination.ResolutionExact ||
		receipt.Repository != input.Repository || receipt.Subject != input.Subject ||
		receipt.Producer != termination.Producer || receipt.Consumer != termination.Consumer ||
		receipt.MetaOperation != termination.MetaOperation || receipt.ProofChoice != termination.ProofChoice ||
		receipt.Stage != termination.TraceStage || receipt.Decision != class.decision ||
		receipt.Reason != class.reason || receipt.InputDigest != digestJSON(input) ||
		receipt.TraceDigest != digestJSON(input.Trace) || !reflect.DeepEqual(receipt.Observations, input.Trace) ||
		!reflect.DeepEqual(receipt.ClaimTransitions, transitions(class)) || receipt.Authority != (termination.Authority{ReadOnly: true}) {
		return fmt.Errorf("independent judge: receipt does not match the observed trace")
	}
	if receipt.Summary.ObservedSteps != len(input.Trace) || receipt.Summary.MaxSteps != input.MaxSteps ||
		receipt.Summary.StateCount != class.stateCount || receipt.Summary.DetectedPeriod != class.period ||
		receipt.Summary.FinalState != class.finalState ||
		receipt.Summary.TerminationProven != (class.decision == termination.DecisionFixedPoint) ||
		receipt.Summary.Total != termination.IndicatorTotal || receipt.Summary.Satisfied != termination.IndicatorTotal ||
		receipt.Summary.BasisPoints != 10000 || len(receipt.Indicators) != termination.IndicatorTotal {
		return fmt.Errorf("independent judge: fixed denominator or summary mismatch")
	}
	for index, indicator := range receipt.Indicators {
		if indicator.ID != indicatorIDs[index] || !indicator.Satisfied || indicator.Route != "TERMINATION" ||
			indicator.Producer != termination.Producer || indicator.Consumer != termination.Consumer ||
			indicator.MetaOperation != termination.MetaOperation || indicator.ProofChoice != termination.ProofChoice ||
			indicator.Stage != termination.ClaimStage || indicator.Step != 0 || indicator.Value != "true" ||
			indicator.Limit != "true" || indicator.Reason != indicatorReasons[index] {
			return fmt.Errorf("independent judge: indicator %d is not bound", index+1)
		}
	}
	if !validDigest(receipt.ReceiptDigest) || receipt.ReceiptDigest != receiptDigest(receipt) ||
		!validDigest(receipt.ReplayDigest) || receipt.ReplayDigest != replayDigest(receipt) {
		return fmt.Errorf("independent judge: receipt digest mismatch")
	}
	return nil
}

var indicatorIDs = []string{
	"gooo.termination.input-schema.v1", "gooo.termination.identity-bound.v1",
	"gooo.termination.bounded-trace.v1", "gooo.termination.state-chain.v1",
	"gooo.termination.no-change-branch.v1", "gooo.termination.cycle-branch.v1",
	"gooo.termination.divergence-branch.v1", "gooo.termination.progress-branch.v1",
	"gooo.termination.claim-transition.v1", "gooo.termination.read-only-authority.v1",
}

var indicatorReasons = []string{
	"INPUT_SCHEMA_BOUND", "PRODUCER_CONSUMER_BOUND", "TRACE_WITHIN_FIXED_BUDGET",
	"CONTIGUOUS_STATE_CHAIN", "NO_CHANGE_BRANCH_EXACT", "CYCLE_BRANCH_EXACT",
	"DIVERGENCE_IS_ONLY_POSSIBLE", "IN_PROGRESS_BRANCH_EXACT", "CLAIM_TRANSITION_BOUND",
	"READ_ONLY_NO_PROMOTION",
}

func classify(trace []termination.Observation, maxSteps int) classification {
	states := []string{trace[0].BeforeState}
	for _, observation := range trace {
		if observation.AfterState != states[len(states)-1] {
			states = append(states, observation.AfterState)
		}
	}
	_, period := repeatedState(states)
	final := trace[len(trace)-1]
	hasCycle := period != 0
	if hasCycle {
		return classification{termination.DecisionCycle, "REPEATED_STATE_CYCLE_OBSERVED", final.AfterState, period, len(states), true, false}
	}
	if final.Decision == "NO_CHANGE" {
		return classification{termination.DecisionFixedPoint, "NO_CHANGE_FIXED_POINT_OBSERVED", final.AfterState, 0, len(states), false, false}
	}
	diverging := len(trace) == maxSteps
	for _, observation := range trace {
		diverging = diverging && observation.Decision == "CHANGED" && observation.AfterRank > observation.BeforeRank
	}
	if diverging {
		return classification{termination.DecisionDivergence, "STRICTLY_GROWING_BOUNDARY_NO_FIXED_POINT", final.AfterState, 0, len(states), false, true}
	}
	return classification{termination.DecisionInProgress, "TRACE_ENDED_BEFORE_TERMINATION", final.AfterState, 0, len(states), false, false}
}

func repeatedState(states []string) (int, int) {
	for left := 0; left < len(states); left++ {
		for right := left + 1; right < len(states); right++ {
			if states[left] == states[right] {
				return left, right - left
			}
		}
	}
	return -1, 0
}

func transitions(class classification) []termination.ClaimTransition {
	return []termination.ClaimTransition{
		{Stage: termination.ClaimStage, Step: 0, From: "UNPROVEN", To: "OBSERVED", Reason: "TRACE_BOUND"},
		{Stage: termination.ClaimStage, Step: class.stateCount - 1, From: "OBSERVED", To: class.decision, Reason: class.reason},
	}
}

func digestJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func receiptDigest(receipt termination.Receipt) string {
	receipt.ReceiptDigest, receipt.ReplayDigest = "", ""
	return digestJSON(receipt)
}

func replayDigest(receipt termination.Receipt) string {
	return digestJSON(struct {
		InputDigest, TraceDigest, ReceiptDigest string
	}{receipt.InputDigest, receipt.TraceDigest, receipt.ReceiptDigest})
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
