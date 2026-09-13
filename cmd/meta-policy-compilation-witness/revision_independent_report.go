package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

type revisionConsumerView struct {
	Schema string `json:"schema"`
	SourceDigest string `json:"source_digest"`
	RequestArtifactDigest string `json:"request_artifact_digest"`
	CanonicalRequestDigest string `json:"canonical_request_digest"`
	ReportArtifactDigest string `json:"report_artifact_digest"`
	Decision string `json:"decision"`
	Checks []struct {
		ID string `json:"id"`
		State string `json:"state"`
	} `json:"checks"`
	Comparisons int `json:"independent_result_comparisons"`
	ExecutionObserved bool `json:"policy_execution_observed"`
	Improvement string `json:"improvement"`
	MutationAuthority int `json:"mutation_authority"`
	PromotionAuthority int `json:"promotion_authority"`
}

func checkRevisionConsumerReport(data []byte, operation policycompilation.PolicyRevisionOperationObservation, report []byte) (string, error) {
	if err := revisionConsumerJSON(data); err != nil {
		return "UNKNOWN", err
	}
	var fields map[string]json.RawMessage
	var view revisionConsumerView
	if err := json.Unmarshal(data, &fields); err != nil {
		return "UNKNOWN", err
	}
	for _, name := range []string{"schema", "source_digest", "request_artifact_digest", "canonical_request_digest",
		"report_artifact_digest", "decision", "checks", "independent_result_comparisons", "policy_execution_observed",
		"improvement", "mutation_authority", "promotion_authority"} {
		if value, present := fields[name]; !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return "UNKNOWN", fmt.Errorf("consumer report requires non-null %s", name)
		}
	}
	if err := json.Unmarshal(data, &view); err != nil {
		return "UNKNOWN", err
	}
	if operation.Observation == nil || view.Schema != "gooo/meta-policy-revision-receipt-observation/v1" ||
		view.SourceDigest != operation.PolicySourceDigest || view.RequestArtifactDigest != operation.RequestArtifactDigest ||
		view.CanonicalRequestDigest != operation.Observation.RequestDigest ||
		view.ReportArtifactDigest != policycompilation.DigestBytes(report) || view.ExecutionObserved ||
		view.Improvement != "UNKNOWN" || view.MutationAuthority != 0 || view.PromotionAuthority != 0 {
		return "UNKNOWN", errors.New("consumer report schema, exact input binding, or authority differs")
	}
	if view.Decision == "REFUTED" {
		return "REFUTED", errors.New("independent consumer refuted the fresh report")
	}
	if view.Decision != "RECEIPT_CONSISTENT_ONLY" {
		return "UNKNOWN", errors.New("consumer did not establish receipt consistency")
	}
	expected := map[string]bool{
		"SOURCE_BINDING": false, "REQUEST_BINDING": false, "REVISION_SCOPE": false,
		"DECLARED_INPUT_PRESERVATION": false, "RESULT_RECONSTRUCTION": false, "ACCOUNTING_RECONSTRUCTION": false,
	}
	for _, check := range view.Checks {
		seen, present := expected[check.ID]
		if !present || seen || check.State != "CLOSED" {
			return "UNKNOWN", errors.New("consumer check set is incomplete, duplicated, unknown, or not closed")
		}
		expected[check.ID] = true
	}
	for _, seen := range expected {
		if !seen {
			return "UNKNOWN", errors.New("consumer omitted an existing receipt check")
		}
	}
	comparisons := 0
	for _, phase := range []policycompilation.PolicyRevisionExecution{operation.Observation.Baseline, operation.Observation.Candidate} {
		comparisons += len(phase.SourceResults) + len(phase.FirstResults) + len(phase.ReplayResults)
	}
	if comparisons == 0 || view.Comparisons != comparisons {
		return "UNKNOWN", errors.New("consumer result comparison count differs from the fresh report")
	}
	return "INDEPENDENT_RECONSTRUCTION_OBSERVED", nil
}

// Preserve one-object framing and reject duplicate keys at every nesting level.
func revisionConsumerJSON(data []byte) error {
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || trimmed[0] != '{' {
		return errors.New("consumer report must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var visit func(int) error
	visit = func(depth int) error {
		if depth > 128 {
			return errors.New("consumer report exceeds JSON nesting bound")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if token == json.Delim('{') {
			seen := map[string]bool{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] {
					return errors.New("consumer report contains a duplicate or invalid key")
				}
				seen[name] = true
				if err := visit(depth + 1); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
		} else if token == json.Delim('[') {
			for decoder.More() {
				if err := visit(depth + 1); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
		}
		return err
	}
	if err := visit(0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("consumer report contains trailing JSON")
	}
	return nil
}
