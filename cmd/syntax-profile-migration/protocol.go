package main

import (
	"encoding/json"
	"io"
	"io/fs"
	"runtime"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func inspectProfileMigration(repository fs.FS, request syntaxregistration.ProfileMigrationRequest,
	after []byte, output io.Writer) error {
	snapshot, source, replacement, err := syntaxregistration.InspectProfileMigrationInputs(repository, request, after)
	if err != nil {
		return err
	}
	identity, err := syntaxregistration.ObserveExecutionIdentity()
	if err != nil {
		return err
	}
	request.SnapshotDigest, request.SourceDigest, request.ReplacementDigest = snapshot, source, replacement
	request.Toolchain, request.ExecutionIdentity = runtime.Version(), identity
	return json.NewEncoder(output).Encode(map[string]any{
		"request":         request,
		"snapshot_digest": snapshot, "source_digest": source, "replacement_digest": replacement,
		"toolchain": runtime.Version(), "execution_identity": identity, "semantic_admission": "UNASSESSED",
		"repository_writes": 0, "apply_authorized": false, "promotion_authorized": false})
}

func verifyProfileCandidate(repository fs.FS, plan syntaxregistration.ProfileMigrationPlan,
	path string, output io.Writer) error {
	raw, err := readProfileFile(path, 16<<20)
	if err != nil {
		return err
	}
	candidate, err := syntaxregistration.DecodeProfileMigrationCandidate(raw)
	if err != nil {
		return err
	}
	if err := plan.ValidateCandidate(repository, candidate); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(map[string]any{
		"status": "EXACT_SOURCE_CANDIDATE", "input_digest": candidate.InputDigest,
		"request_digest": candidate.RequestDigest, "semantic_admission": candidate.Admission,
		"repository_writes": 0, "apply_authorized": false, "promotion_authorized": false})
}
