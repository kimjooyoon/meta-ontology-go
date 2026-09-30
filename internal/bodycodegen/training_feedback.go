package bodycodegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// ExternalTrainingFeedback carries advisory observations from an earlier,
// source-bound training run. Its claims are not authenticated execution proof
// and do not replace this invocation's candidate evaluation.
type ExternalTrainingFeedback struct {
	SourceDigest        string                 `json:"source_digest"`
	TrainingSuiteSHA256 string                 `json:"training_suite_sha256"`
	CandidateID         string                 `json:"candidate_id"`
	Observations        []IRBodyFillCaseResult `json:"observations"`
}

// UnmarshalJSON requires every observation field to be present and non-null.
// The typed result remains compact and shared with local search receipts, while
// the wire check prevents absent integer/bool fields from silently becoming 0.
func (feedback *ExternalTrainingFeedback) UnmarshalJSON(data []byte) error {
	var wire struct {
		SourceDigest        string            `json:"source_digest"`
		TrainingSuiteSHA256 string            `json:"training_suite_sha256"`
		CandidateID         string            `json:"candidate_id"`
		Observations        []json.RawMessage `json:"observations"`
	}
	if err := decodeStrictJSON(data, &wire); err != nil {
		return err
	}
	observations := make([]IRBodyFillCaseResult, 0, len(wire.Observations))
	for index, raw := range wire.Observations {
		var item struct {
			Input    *int64 `json:"input"`
			Expected *int64 `json:"expected"`
			Actual   *int64 `json:"actual"`
			Passed   *bool  `json:"passed"`
		}
		if err := decodeStrictJSON(raw, &item); err != nil {
			return fmt.Errorf("external training observation %d: %w", index, err)
		}
		if item.Input == nil || item.Expected == nil || item.Actual == nil || item.Passed == nil {
			return fmt.Errorf("external training observation %d requires explicit non-null input, expected, actual, and passed fields", index)
		}
		observations = append(observations, IRBodyFillCaseResult{
			Input: *item.Input, Expected: *item.Expected, Actual: *item.Actual, Passed: *item.Passed,
		})
	}
	*feedback = ExternalTrainingFeedback{
		SourceDigest: wire.SourceDigest, TrainingSuiteSHA256: wire.TrainingSuiteSHA256,
		CandidateID: wire.CandidateID, Observations: observations,
	}
	return nil
}

func decodeStrictJSON(data []byte, target any) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("JSON must be valid UTF-8")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := parseJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func parseJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		return parseJSONObject(decoder)
	case '[':
		return parseJSONArray(decoder)
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}

func parseJSONObject(decoder *json.Decoder) error {
	seen := []string{}
	for decoder.More() {
		if err := parseJSONObjectMember(decoder, &seen); err != nil {
			return err
		}
	}
	return consumeJSONDelimiter(decoder, '}')
}

func parseJSONObjectMember(decoder *json.Decoder, seen *[]string) error {
	keyToken, err := decoder.Token()
	if err != nil {
		return err
	}
	key, ok := keyToken.(string)
	if !ok {
		return fmt.Errorf("JSON object key is not a string")
	}
	if err := addUniqueJSONKey(seen, key); err != nil {
		return err
	}
	return parseJSONValue(decoder)
}

func addUniqueJSONKey(seen *[]string, key string) error {
	for _, prior := range *seen {
		if strings.EqualFold(prior, key) {
			return fmt.Errorf("duplicate JSON field %q (case-insensitive match for %q)", key, prior)
		}
	}
	*seen = append(*seen, key)
	return nil
}

func parseJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := parseJSONValue(decoder); err != nil {
			return err
		}
	}
	return consumeJSONDelimiter(decoder, ']')
}

func consumeJSONDelimiter(decoder *json.Decoder, expected json.Delim) error {
	end, err := decoder.Token()
	if err != nil {
		return err
	}
	if end != expected {
		if expected == '}' {
			return fmt.Errorf("unterminated JSON object")
		}
		return fmt.Errorf("unterminated JSON array")
	}
	return nil
}

// IRBodySearchExternalFeedbackReceipt records local provenance bindings and
// counts without elevating external observations to verified execution.
type IRBodySearchExternalFeedbackReceipt struct {
	FeedbackSHA256              string `json:"feedback_sha256"`
	SourceDigest                string `json:"source_digest"`
	TrainingSuiteSHA256         string `json:"training_suite_sha256"`
	CandidateID                 string `json:"candidate_id"`
	ObservationCount            int    `json:"observation_count"`
	FailedObservationCount      int    `json:"failed_observation_count"`
	PromptedFailureCount        int    `json:"prompted_failure_count"`
	FailedObservationsTruncated bool   `json:"failed_observations_truncated"`
}

type bodySearchExternalFailure struct {
	Input    int64 `json:"input"`
	Expected int64 `json:"expected"`
	Actual   int64 `json:"actual"`
}

// bodySearchExternalFeedbackPrompt is intentionally smaller than the source
// artifact: no digest strings or claimed pass rows reach model state.
type bodySearchExternalFeedbackPrompt struct {
	CandidateID          string                      `json:"candidate_id"`
	FailedCases          []bodySearchExternalFailure `json:"failed_cases"`
	FailedCasesTotal     int                         `json:"failed_cases_total"`
	FailedCasesTruncated bool                        `json:"failed_cases_truncated"`
}

func validateExternalTrainingFeedback(plan IRBodySearchPlan, source []byte) error {
	feedback := plan.ExternalTrainingFeedback
	if feedback == nil {
		return nil
	}
	if feedback.SourceDigest != digest(source) {
		return fmt.Errorf("external training feedback source_digest does not match the original DSL source")
	}
	trainingBytes, err := json.Marshal(plan.TestCases)
	if err != nil {
		return fmt.Errorf("encode canonical training cases: %w", err)
	}
	if feedback.TrainingSuiteSHA256 != digest(trainingBytes) {
		return fmt.Errorf("external training feedback training_suite_sha256 does not match the canonical typed training cases")
	}
	return nil
}

func externalTrainingFeedbackReceipt(feedback *ExternalTrainingFeedback) *IRBodySearchExternalFeedbackReceipt {
	if feedback == nil {
		return nil
	}
	encoded, _ := json.Marshal(feedback)
	failed := 0
	for _, observation := range feedback.Observations {
		if !observation.Passed {
			failed++
		}
	}
	return &IRBodySearchExternalFeedbackReceipt{
		FeedbackSHA256: digest(encoded), SourceDigest: feedback.SourceDigest,
		TrainingSuiteSHA256: feedback.TrainingSuiteSHA256, CandidateID: feedback.CandidateID,
		ObservationCount: len(feedback.Observations), FailedObservationCount: failed,
		PromptedFailureCount: min(failed, 8), FailedObservationsTruncated: failed > 8,
	}
}

func externalTrainingFeedbackPrompt(feedback *ExternalTrainingFeedback) *bodySearchExternalFeedbackPrompt {
	if feedback == nil {
		return nil
	}
	prompt := &bodySearchExternalFeedbackPrompt{
		CandidateID: feedback.CandidateID, FailedCases: []bodySearchExternalFailure{},
	}
	for _, observation := range feedback.Observations {
		if observation.Passed {
			continue
		}
		prompt.FailedCasesTotal++
		if len(prompt.FailedCases) < 8 {
			prompt.FailedCases = append(prompt.FailedCases, bodySearchExternalFailure{
				Input: observation.Input, Expected: observation.Expected, Actual: observation.Actual,
			})
		}
	}
	prompt.FailedCasesTruncated = prompt.FailedCasesTotal > len(prompt.FailedCases)
	return prompt
}
