package publicorchestration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const (
	HandoffSchema = "gooo/public-self-improvement-orchestration-handoff/v1"
	ReportSchema  = "gooo/public-self-improvement-orchestration-report/v1"
)

type Handoff struct {
	Schema                  string        `json:"schema"`
	HandoffID               string        `json:"handoff_id"`
	Operation               string        `json:"operation"`
	Decision                string        `json:"decision"`
	State                   string        `json:"state"`
	NextOperation           string        `json:"next_operation"`
	Unknown                 *UnknownState `json:"unknown"`
	PolicySourceDigest      string        `json:"policy_source_digest"`
	PolicySemanticDigest    string        `json:"policy_semantic_digest"`
	PolicyEvaluatorDigest   string        `json:"policy_evaluator_digest"`
	CandidateDigest         string        `json:"candidate_digest"`
	CandidateID             string        `json:"candidate_id"`
	GeneratedOutputDigest   string        `json:"generated_output_digest"`
	GeneratedManifestDigest string        `json:"generated_manifest_digest"`
	RequiredArtifacts       []string      `json:"required_artifacts"`
	AuthorizationRequired   bool          `json:"authorization_required"`
	ExecutionAllowed        bool          `json:"execution_allowed"`
	RepositoryWrites        int           `json:"repository_writes"`
	LocalTestExecutions     int           `json:"local_test_executions"`
}

func HandoffContentDigest(handoff Handoff) (string, error) {
	handoff.HandoffID = ""
	digest, err := cache.DigestOf(handoff)
	if err != nil {
		return "", fmt.Errorf("orchestration handoff content digest: %w", err)
	}
	return digest.String(), nil
}

func ValidateHandoff(handoff Handoff, policy Policy) error {
	if handoff.Schema != HandoffSchema || !cache.Digest(handoff.HandoffID).Known() || handoff.Operation != policy.Operation ||
		handoff.Decision != DecisionUnknown || handoff.State != policy.Boundary || handoff.NextOperation == "" || handoff.Unknown == nil ||
		!handoff.AuthorizationRequired || handoff.ExecutionAllowed || handoff.RepositoryWrites != 0 || handoff.LocalTestExecutions != 0 ||
		!sameValues(handoff.RequiredArtifacts, policy.HandoffArtifacts) {
		return errors.New("orchestration handoff identity is invalid")
	}
	for _, value := range []string{handoff.PolicySourceDigest, handoff.PolicySemanticDigest, handoff.PolicyEvaluatorDigest, handoff.CandidateDigest, handoff.GeneratedOutputDigest, handoff.GeneratedManifestDigest} {
		if !cache.Digest(value).Known() {
			return errors.New("orchestration handoff contains unknown digest evidence")
		}
	}
	if handoff.PolicySourceDigest != policy.SourceDigest || handoff.PolicySemanticDigest != policy.SemanticDigest || handoff.PolicyEvaluatorDigest != policy.EvaluatorDigest || handoff.Unknown.NextOperation != handoff.NextOperation || !SameUnknown(handoff.Unknown, policy.UnknownFor(CaseMissingAuthorization)) {
		return errors.New("orchestration handoff does not bind the exact lowered policy")
	}
	digest, err := HandoffContentDigest(handoff)
	if err != nil || digest != handoff.HandoffID {
		return errors.New("orchestration handoff is not content-addressed")
	}
	return nil
}

func MarshalHandoff(handoff Handoff, policy Policy) ([]byte, error) {
	if err := ValidateHandoff(handoff, policy); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(handoff, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func DecodeHandoff(data []byte, policy Policy) (Handoff, error) {
	var handoff Handoff
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&handoff); err != nil {
		return Handoff{}, fmt.Errorf("decode orchestration handoff: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Handoff{}, errors.New("orchestration handoff contains multiple JSON values")
	} else if err != io.EOF {
		return Handoff{}, fmt.Errorf("decode orchestration handoff trailer: %w", err)
	}
	if err := ValidateHandoff(handoff, policy); err != nil {
		return Handoff{}, err
	}
	return handoff, nil
}

func WriteHandoff(filename string, handoff Handoff, policy Policy) error {
	data, err := MarshalHandoff(handoff, policy)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o444)
	if err != nil {
		return fmt.Errorf("create immutable orchestration handoff: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func ReadHandoff(filename string, policy Policy) (Handoff, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return Handoff{}, err
	}
	if !info.Mode().IsRegular() {
		return Handoff{}, errors.New("orchestration handoff is not a regular file")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return Handoff{}, err
	}
	return DecodeHandoff(data, policy)
}

func NewUnknown(policy Policy, caseID string) (*UnknownState, error) {
	unknown := policy.UnknownFor(caseID)
	if unknown == nil {
		return nil, fmt.Errorf("orchestration case %q is not UNKNOWN", caseID)
	}
	return unknown, nil
}

func SameUnknown(left, right *UnknownState) bool {
	if left == nil || right == nil {
		return left == right
	}
	if left.Stage != right.Stage || left.Step != right.Step || left.Reason != right.Reason || left.UnknownClass != right.UnknownClass || left.NextOperation != right.NextOperation {
		return false
	}
	return sameValues(left.BlockedBy, right.BlockedBy)
}
