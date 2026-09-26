package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const JEVDecisionBoundarySchemaPart01 = "gooo/generation/jev-decision-boundary/v1"

type JEVDecisionStatusPart01 string

const (
	JEVDecisionPassPart01    JEVDecisionStatusPart01 = "PASS"
	JEVDecisionUnknownPart01 JEVDecisionStatusPart01 = "UNKNOWN"
)

type JEVDecisionKindPart01 string

const (
	JEVDecisionChoicePart01 JEVDecisionKindPart01 = "choice"
	JEVDecisionScorePart01  JEVDecisionKindPart01 = "score"
	JEVDecisionNoulPart01   JEVDecisionKindPart01 = "noul"
)

type JEVDecisionBoundaryInputPart01 struct {
	StateDigest             string                  `json:"state_digest"`
	QuestionDigest          string                  `json:"question_digest"`
	ModelIdentity           string                  `json:"model_identity"`
	PolicyIdentity          string                  `json:"policy_identity"`
	ResponseDigest          string                  `json:"response_digest"`
	ActionObservationDigest string                  `json:"action_observation_digest,omitempty"`
	DecisionKind            JEVDecisionKindPart01   `json:"decision_kind"`
	Decision                string                  `json:"decision"`
	Status                  JEVDecisionStatusPart01 `json:"status"`
	Reason                  string                  `json:"reason,omitempty"`
	Confidence              float64                 `json:"confidence"`
	MissingStageIndex       int                     `json:"missing_stage_index"`
}

type JEVDecisionBoundaryPart01 struct {
	Schema                  string                  `json:"schema"`
	StateDigest             string                  `json:"state_digest"`
	QuestionDigest          string                  `json:"question_digest"`
	ModelIdentity           string                  `json:"model_identity"`
	PolicyIdentity          string                  `json:"policy_identity"`
	ResponseDigest          string                  `json:"response_digest"`
	ActionObservationDigest string                  `json:"action_observation_digest,omitempty"`
	DecisionKind            JEVDecisionKindPart01   `json:"decision_kind"`
	Decision                string                  `json:"decision"`
	Status                  JEVDecisionStatusPart01 `json:"status"`
	Reason                  string                  `json:"reason,omitempty"`
	Confidence              float64                 `json:"confidence"`
	MissingStageIndex       int                     `json:"missing_stage_index"`
	EvidencePrefixDigest    string                  `json:"evidence_prefix_digest"`
	DecisionDigest          string                  `json:"decision_digest"`
}

func NewJEVDecisionBoundaryPart01(input JEVDecisionBoundaryInputPart01) (JEVDecisionBoundaryPart01, error) {
	receipt := JEVDecisionBoundaryPart01{
		Schema:                  JEVDecisionBoundarySchemaPart01,
		StateDigest:             input.StateDigest,
		QuestionDigest:          input.QuestionDigest,
		ModelIdentity:           input.ModelIdentity,
		PolicyIdentity:          input.PolicyIdentity,
		ResponseDigest:          input.ResponseDigest,
		ActionObservationDigest: input.ActionObservationDigest,
		DecisionKind:            input.DecisionKind,
		Decision:                input.Decision,
		Status:                  input.Status,
		Reason:                  input.Reason,
		Confidence:              input.Confidence,
		MissingStageIndex:       input.MissingStageIndex,
	}
	if err := receipt.validateShapePart01(); err != nil {
		return JEVDecisionBoundaryPart01{}, err
	}
	receipt.EvidencePrefixDigest = digestJEVPart01(receipt.evidencePrefixPart01())
	receipt.DecisionDigest = digestJEVPart01(receipt.decisionViewPart01())
	return receipt, nil
}

func (r JEVDecisionBoundaryPart01) ValidPart01() bool {
	if r.validateShapePart01() != nil || !validDigestJEVPart01(r.EvidencePrefixDigest) || !validDigestJEVPart01(r.DecisionDigest) {
		return false
	}
	if r.EvidencePrefixDigest != digestJEVPart01(r.evidencePrefixPart01()) {
		return false
	}
	return r.DecisionDigest == digestJEVPart01(r.decisionViewPart01())
}

func (r JEVDecisionBoundaryPart01) validateShapePart01() error {
	if r.Schema != JEVDecisionBoundarySchemaPart01 {
		return errors.New("invalid JEV decision boundary schema")
	}
	for name, value := range map[string]string{
		"state digest":    r.StateDigest,
		"question digest": r.QuestionDigest,
		"model identity":  r.ModelIdentity,
		"policy identity": r.PolicyIdentity,
		"response digest": r.ResponseDigest,
	} {
		if !validDigestJEVPart01(value) && name != "model identity" && name != "policy identity" {
			return fmt.Errorf("invalid %s", name)
		}
		if (name == "model identity" || name == "policy identity") && value == "" {
			return fmt.Errorf("missing %s", name)
		}
	}
	switch r.DecisionKind {
	case JEVDecisionChoicePart01, JEVDecisionScorePart01, JEVDecisionNoulPart01:
	default:
		return errors.New("invalid JEV decision kind")
	}
	switch r.Status {
	case JEVDecisionPassPart01:
		if r.MissingStageIndex != -1 {
			return errors.New("PASS requires missing stage index -1")
		}
	case JEVDecisionUnknownPart01:
		if r.MissingStageIndex < 0 || r.Reason == "" {
			return errors.New("UNKNOWN requires reason and missing stage index")
		}
	default:
		return errors.New("invalid JEV decision status")
	}
	if r.Decision == "" || r.Confidence < 0 || r.Confidence > 1 {
		return errors.New("invalid JEV decision")
	}
	if r.ActionObservationDigest != "" && !validDigestJEVPart01(r.ActionObservationDigest) {
		return errors.New("invalid action observation digest")
	}
	return nil
}

func (r JEVDecisionBoundaryPart01) evidencePrefixPart01() any {
	return struct {
		Schema                  string                  `json:"schema"`
		StateDigest             string                  `json:"state_digest"`
		QuestionDigest          string                  `json:"question_digest"`
		ModelIdentity           string                  `json:"model_identity"`
		PolicyIdentity          string                  `json:"policy_identity"`
		ResponseDigest          string                  `json:"response_digest"`
		ActionObservationDigest string                  `json:"action_observation_digest,omitempty"`
		DecisionKind            JEVDecisionKindPart01   `json:"decision_kind"`
		Decision                string                  `json:"decision"`
		Status                  JEVDecisionStatusPart01 `json:"status"`
		Reason                  string                  `json:"reason,omitempty"`
		Confidence              float64                 `json:"confidence"`
		MissingStageIndex       int                     `json:"missing_stage_index"`
	}{r.Schema, r.StateDigest, r.QuestionDigest, r.ModelIdentity, r.PolicyIdentity, r.ResponseDigest, r.ActionObservationDigest, r.DecisionKind, r.Decision, r.Status, r.Reason, r.Confidence, r.MissingStageIndex}
}

func (r JEVDecisionBoundaryPart01) decisionViewPart01() any {
	return struct {
		Prefix string `json:"evidence_prefix_digest"`
		Input  any    `json:"input"`
	}{r.EvidencePrefixDigest, r.evidencePrefixPart01()}
}

func digestJEVPart01(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigestJEVPart01(value string) bool {
	if len(value) != len("sha256:")+64 || len(value) < 7 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
