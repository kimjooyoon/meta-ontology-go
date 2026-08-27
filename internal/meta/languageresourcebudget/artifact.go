package languageresourcebudget

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type sourceReceipt struct {
	SchemaVersion string            `json:"schema_version"`
	Command       string            `json:"command"`
	Status        string            `json:"status"`
	File          string            `json:"file"`
	Diagnostics   []json.RawMessage `json:"diagnostics"`
}

type artifact struct {
	Schema        string `json:"schema"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	Kind          string `json:"kind"`
	SubjectDigest string `json:"subject_digest"`
	Operation     struct {
		Activity string `json:"activity"`
	} `json:"operation"`
	Effects Effects `json:"effects"`
	Digest  string  `json:"digest"`
}

func verifyProducer(input Input) (Semantic, error) {
	var source sourceReceipt
	if err := json.Unmarshal(input.Producer.SourceReceipt, &source); err != nil {
		return Semantic{}, fmt.Errorf("SOURCE_RECEIPT_INVALID")
	}
	if source.SchemaVersion != "gooo/diagnostics/v1" || source.Command != "check" || source.Status != "ok" ||
		source.File != input.Contract.SourcePaths[0] || len(source.Diagnostics) != 0 {
		return Semantic{}, fmt.Errorf("SOURCE_RECEIPT_NOT_EXACT")
	}
	var first, replay artifact
	if err := json.Unmarshal(input.Producer.Artifact, &first); err != nil {
		return Semantic{}, fmt.Errorf("ARTIFACT_INVALID")
	}
	if err := json.Unmarshal(input.Producer.Replay, &replay); err != nil {
		return Semantic{}, fmt.Errorf("REPLAY_INVALID")
	}
	valid := func(value artifact) bool {
		return value.Schema == "gooo/operation-manifest/v1" && value.Decision == "PASS" && value.Resolution == "EXACT" &&
			value.Reason == "OPERATION_MANIFEST_EMITTED" && value.Kind == "operation-manifest" &&
			value.Operation.Activity == input.Contract.Entry && value.Effects.RepositoryWrites == 0 &&
			!value.Effects.MutationAuthority && len(value.SubjectDigest) > 0 && len(value.Digest) > 0
	}
	if !valid(first) || !valid(replay) {
		return Semantic{}, fmt.Errorf("ARTIFACT_SEMANTICS_INVALID")
	}
	firstDigest, replayDigest := digestBytes(input.Producer.Artifact), digestBytes(input.Producer.Replay)
	if !bytes.Equal(input.Producer.Artifact, input.Producer.Replay) {
		return Semantic{Decision: "FAIL_CLOSED", Resolution: "EXACT", Reason: "ARTIFACT_REPLAY_MISMATCH", SourceDigest: input.Producer.SourceDigest, ArtifactDigest: firstDigest, ReplayDigest: replayDigest}, fmt.Errorf("ARTIFACT_REPLAY_MISMATCH")
	}
	return Semantic{Decision: "PASS", Resolution: "EXACT", Reason: "SEMANTIC_ARTIFACT_REPLAY_STABLE", SourceDigest: input.Producer.SourceDigest, ArtifactDigest: firstDigest, ReplayDigest: replayDigest}, nil
}
