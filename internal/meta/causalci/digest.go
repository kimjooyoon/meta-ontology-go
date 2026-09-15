package causalci

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ClaimInstanceID is stable for one template/proposition/subject tuple and
// prevents one template ID from being reused as the identity of many claims.
func ClaimInstanceID(templateID, subjectPath, proposition string) string {
	digest, _ := digestJSON(struct {
		TemplateID  string `json:"template_id"`
		SubjectPath string `json:"subject_path"`
		Proposition string `json:"proposition"`
	}{templateID, subjectPath, proposition})
	return "claim-instance:" + strings.TrimPrefix(digest, "sha256:")
}

// GitBlobObjectID reproduces git's blob object coordinate without consulting
// the repository. It lets both producer and consumer detect source mutation
// between observation and reconstruction.
func GitBlobObjectID(data []byte) string {
	return GitBlobObjectIDForFormat(data, "sha1")
}

func GitBlobObjectIDForFormat(data []byte, format string) string {
	header := []byte(fmt.Sprintf("blob %d\x00", len(data)))
	payload := append(header, data...)
	if strings.EqualFold(format, "sha256") {
		hash := sha256.Sum256(payload)
		return hex.EncodeToString(hash[:])
	}
	hash := sha1.Sum(payload)
	return hex.EncodeToString(hash[:])
}

func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func transitionDigest(value ClaimTransition) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func receiptDigest(value Receipt) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func planDigest(value Receipt) (string, error) {
	projection := struct {
		ObservationDigest string                `json:"observation_digest"`
		Operation         Operation             `json:"operation"`
		ExecutionMode     string                `json:"execution_mode"`
		Conformance       Conformance           `json:"conformance"`
		PlanGate          PlanGate              `json:"plan_gate"`
		Contradictions    []PolicyContradiction `json:"policy_contradictions"`
		Subjects          []SubjectResolution   `json:"subjects"`
		ClaimTransitions  []ClaimTransition     `json:"claim_transitions"`
	}{
		ObservationDigest: value.ObservationDigest,
		Operation:         value.Operation,
		ExecutionMode:     value.ExecutionMode,
		Conformance:       value.Conformance,
		PlanGate:          value.PlanGate,
		Contradictions:    value.PolicyContradictions,
		Subjects:          value.Subjects,
		ClaimTransitions:  value.ClaimTransitions,
	}
	return digestJSON(projection)
}

func interventionDigest(value InterventionReport) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
