package claimledger

import (
	"encoding/json"
	"strings"
)

func validateRuntimeSource(source, observation sourceState, subject string) sourceState {
	if source.Status != "VERIFIED" {
		return source
	}
	if observation.Status != "VERIFIED" {
		source.Status, source.Reason = "INVALID", "RUNTIME_EVIDENCE_OBSERVATION_SOURCE_INVALID"
		return source
	}
	checks := []struct {
		path   string
		want   string
		reason string
	}{
		{"schema", "gooo/runtime-measurement-evidence/v1", "RUNTIME_EVIDENCE_SCHEMA_MISMATCH"},
		{"subject_sha", subject, "RUNTIME_EVIDENCE_SUBJECT_MISMATCH"},
		{"source.observation_digest", observation.Digest, "RUNTIME_EVIDENCE_OBSERVATION_DIGEST_MISMATCH"},
		{"coordinate.stage", "OBSERVE", "RUNTIME_EVIDENCE_STAGE_MISMATCH"},
		{"coordinate.step", "capture-peak-rss", "RUNTIME_EVIDENCE_STEP_MISMATCH"},
		{"producer.tool", "GNU time", "RUNTIME_EVIDENCE_PRODUCER_MISMATCH"},
		{"producer.binary_path", "/usr/bin/time", "RUNTIME_EVIDENCE_PRODUCER_MISMATCH"},
		{"measurement.target", "symbolic-reader-request-observer", "RUNTIME_EVIDENCE_TARGET_MISMATCH"},
		{"measurement.unit", "KiB", "RUNTIME_EVIDENCE_UNIT_MISMATCH"},
	}
	for _, check := range checks {
		value, found := lookup(source.Value, check.path)
		text, valid := value.(string)
		if !found || !valid || text != check.want {
			source.Status, source.Reason = "INVALID", check.reason
			return source
		}
	}
	for _, path := range []string{"producer.binary_digest", "producer.version_digest"} {
		value, found := lookup(source.Value, path)
		text, valid := value.(string)
		if !found || !valid || len(text) != 71 || !strings.HasPrefix(text, "sha256:") {
			source.Status, source.Reason = "INVALID", "RUNTIME_EVIDENCE_PRODUCER_DIGEST_INVALID"
			return source
		}
	}
	writes, writesFound := lookup(source.Value, "effects.repository_writes")
	mutation, mutationFound := lookup(source.Value, "effects.mutation_authority")
	writeNumber, writeValid := writes.(json.Number)
	writeCount, writeError := writeNumber.Int64()
	mutationValue, mutationValid := mutation.(bool)
	if !writesFound || !writeValid || writeError != nil || writeCount != 0 || !mutationFound || !mutationValid || mutationValue {
		source.Status, source.Reason = "INVALID", "RUNTIME_EVIDENCE_EFFECTS_MISMATCH"
		return source
	}
	source.Reason = "SUBJECT_AND_OBSERVATION_BOUND"
	return source
}
