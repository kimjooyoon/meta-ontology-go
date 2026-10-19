package syntaxregistration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// AppendRecord is Gooo's proposed next registration boundary. The native
// denominator reader independently checks the observed baseline and counts.
type AppendRecord struct {
	CaseID         string `json:"case_id"`
	Path           string `json:"path"`
	SourceDigest   string `json:"source_digest"`
	BaselineDigest string `json:"baseline_digest"`
	BaseVersion    int    `json:"base_version"`
	BaseTarget     int    `json:"base_target"`
	CorpusTarget   int    `json:"corpus_target"`
	NextVersion    int    `json:"next_version"`
	NextTarget     int    `json:"next_target"`
	Accepted       bool   `json:"accepted"`
}

type appendObservation struct {
	CaseID         string `json:"case_id"`
	Path           string `json:"path"`
	SourceDigest   string `json:"source_digest"`
	BaselineDigest string `json:"baseline_digest"`
	BaseVersion    int    `json:"base_version"`
	BaseTarget     int    `json:"base_target"`
	CorpusTarget   int    `json:"corpus_target"`
}

func composeRegistrationAppend(ctx context.Context, request Request, previous []byte,
	capability int, source []byte) (AppendRecord, bodycodegen.FixedValue, error) {
	observation, err := observeRegistrationAppend(request, previous, capability)
	if err != nil {
		return AppendRecord{}, bodycodegen.FixedValue{}, err
	}
	input, err := json.Marshal([]any{observation})
	if err != nil {
		return AppendRecord{}, bodycodegen.FixedValue{}, err
	}
	value, err := bodycodegen.EvaluateFixedValue(ctx, "internal/meta/syntaxregistration/contract.gooo",
		source, "ComposeRegistrationAppend", input)
	if err != nil {
		return AppendRecord{}, value, fmt.Errorf("evaluate Gooo registration append: %w", err)
	}
	if value.ActivityID != "gooo://syntax-registration/compose-append" ||
		value.SourceSHA256 != digest(source) || value.ProjectionSHA256 == "" {
		return AppendRecord{}, value, failure("REFUTED", "compose-append",
			"REGISTRATION_APPEND_ORIGIN_MISMATCH", "", "restore-append-source")
	}
	record, err := decodeAppendRecord(value.Value)
	if err != nil {
		return AppendRecord{}, value, err
	}
	if err := checkAppendRecord(record, observation); err != nil {
		return AppendRecord{}, value, err
	}
	return record, value, nil
}

func observeRegistrationAppend(request Request, previous []byte, capability int) (appendObservation, error) {
	var baseline denominator
	if err := decodeStrict(previous, &baseline); err != nil || len(baseline.Boundaries) != len(denominatorBoundaryIDs) {
		return appendObservation{}, failure("REFUTED", "compose-append",
			"REGISTRATION_BASELINE_DENOMINATOR_MISMATCH", "", "restore-bound-baseline")
	}
	return appendObservation{CaseID: request.Case.ID, Path: request.Case.Path,
		SourceDigest: request.SourceDigest, BaselineDigest: digest(previous),
		BaseVersion: request.BaseVersion, BaseTarget: baseline.Boundaries[0].Target,
		CorpusTarget: capability}, nil
}

func decodeAppendRecord(raw []byte) (AppendRecord, error) {
	var record AppendRecord
	if err := decodeStrict(raw, &record); err != nil {
		return AppendRecord{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || len(fields) != 10 {
		return AppendRecord{}, failure("REFUTED", "compose-append",
			"REGISTRATION_APPEND_VALUE_INCOMPLETE", "", "repair-append-activity")
	}
	for _, field := range fields {
		if bytes.Equal(bytes.TrimSpace(field), []byte("null")) {
			return AppendRecord{}, failure("REFUTED", "compose-append",
				"REGISTRATION_APPEND_VALUE_INCOMPLETE", "", "repair-append-activity")
		}
	}
	return record, nil
}

func checkAppendRecord(record AppendRecord, observation appendObservation) error {
	if !record.Accepted {
		return failure("REFUTED", "compose-append",
			"REGISTRATION_APPEND_REFUTED", "", "repair-registration-observation")
	}
	if record.CaseID != observation.CaseID || record.Path != observation.Path ||
		record.SourceDigest != observation.SourceDigest || record.BaselineDigest != observation.BaselineDigest ||
		record.BaseVersion != observation.BaseVersion || record.BaseTarget != observation.BaseTarget ||
		record.CorpusTarget != observation.CorpusTarget || record.NextVersion != observation.BaseVersion+1 ||
		record.NextTarget != observation.CorpusTarget+1 {
		return failure("REFUTED", "compose-append",
			"REGISTRATION_APPEND_CONTRACT_MISMATCH", "", "repair-append-activity")
	}
	return nil
}
