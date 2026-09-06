package main

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

type registrationInputBinding struct {
	Mode            string                        `json:"mode"`
	HeadSHA         string                        `json:"head_sha"`
	HeadObservation generation.ProcessObservation `json:"head_observation"`
}

type registrationScopedEvidence struct {
	Schema    string                   `json:"schema"`
	Input     registrationInputBinding `json:"input"`
	Operation json.RawMessage          `json:"operation"`
}

// A typed registration request can bind a separate, explicit input snapshot.
// Other operations still verify the full common workspace without exclusions.
func registrationBoundInputRoot(workspace, head string) (string, registrationInputBinding, *operationError) {
	binding := registrationInputBinding{Mode: "COMMON_WORKSPACE"}
	if *registrationRoot != "" {
		workspace = *registrationRoot
		binding.Mode = "EXPLICIT_REGISTRATION_SNAPSHOT"
	}
	root, err := filepath.Abs(workspace)
	if err != nil {
		return "", binding, registrationNativeFailure(err, "resolve-registration-input-scope")
	}
	process := registrationRun(root, []string{"git", "rev-parse", "HEAD"}, "git", "rev-parse", "HEAD")
	binding.HeadObservation = process.observation
	if process.err != nil {
		return "", binding, registrationProcessFailure("observe-registration-input-head", process)
	}
	binding.HeadSHA = strings.TrimSpace(string(process.stdout))
	if binding.HeadSHA != head {
		return "", binding, newOperationError("OPERATION_INPUT", "bind-registration-input-head",
			"REGISTRATION_INPUT_HEAD_STALE", "STALE", "restore-exact-registration-snapshot")
	}
	return root, binding, nil
}

// Source digests use the backend's sha256: prefix; semantic IR hashes are bare
// hexadecimal StableHash values. Neither representation is guessed or relaxed.
func registrationContractMatches(candidate syntaxregistration.Candidate, action generation.Action) bool {
	return candidate.ContractDigest == "sha256:"+action.InputContractSourceDigest &&
		candidate.SemanticDigest == action.InputContractSemanticDigest
}

func bindRegistrationInputEvidence(materialized operationMaterialization,
	input registrationInputBinding) (operationMaterialization, *operationError) {
	raw, err := json.Marshal(registrationScopedEvidence{
		Schema: "gooo/native-registration-scoped-instance/v1",
		Input: input, Operation: json.RawMessage(materialized.Canonical)})
	if err != nil {
		return materialized, registrationNativeFailure(err, "encode-input-scope-evidence")
	}
	materialized.Canonical, materialized.InstanceDigest = raw, digestBytes(raw)
	return materialized, nil
}
