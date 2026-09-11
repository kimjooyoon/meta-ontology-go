package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func executeNativeRegistration(workspace string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	var materialized operationMaterialization
	if !generation.ValidRegistrationActionInput(action) {
		return materialized, newOperationError("OPERATION_INPUT", "bind-registration-request",
			"REGISTRATION_ACTION_INPUT_MISMATCH", "KNOWN_CONTRADICTION", "restore-exact-typed-request")
	}
	inputRoot, inputBinding, inputFailure := registrationBoundInputRoot(workspace, plan.HeadSHA)
	if inputFailure != nil {
		return materialized, inputFailure
	}
	workspace = inputRoot
	request := *action.RegistrationRequest
	compiled, err := syntaxregistration.Compile(os.DirFS(workspace), request)
	if err != nil {
		return materialized, registrationNativeFailure(err, "compile-registration")
	}
	temporary, err := os.MkdirTemp("", "gooo-native-registration-")
	if err != nil {
		return materialized, registrationNativeFailure(err, "prepare-output")
	}
	defer os.RemoveAll(temporary)
	requestPath := filepath.Join(temporary, "request.json")
	raw, err := json.Marshal(request)
	if err != nil {
		return materialized, registrationNativeFailure(err, "encode-request")
	}
	if err := os.WriteFile(requestPath, raw, 0600); err != nil {
		return materialized, registrationNativeFailure(err, "write-private-request")
	}
	binary, err := os.Executable()
	if err != nil {
		return materialized, registrationNativeFailure(err, "resolve-worker")
	}
	descriptor := []string{"<meta-execution>", "--registration-mode=worker",
		"--registration-root=<input>", "--registration-request=<request>"}
	args := []string{"--registration-mode=worker", "--registration-root=" + workspace,
		"--registration-request=" + requestPath}
	first := registrationRun(workspace, descriptor, binary, args...)
	materialized.Executor = first.observation
	if first.err != nil {
		return materialized, registrationWorkerFailure(first)
	}
	second := registrationRun(workspace, descriptor, binary, args...)
	materialized.Evaluator = second.observation
	if second.err != nil {
		return materialized, registrationWorkerFailure(second)
	}
	if !bytes.Equal(first.stdout, second.stdout) || !bytes.Equal(first.stderr, second.stderr) {
		return materialized, newOperationError("REPLAY", "compare-registration-workers",
			"REGISTRATION_REPLAY_MISMATCH", "KNOWN_CONTRADICTION", "preserve-replay-counterexample")
	}
	var candidate syntaxregistration.Candidate
	if err := json.Unmarshal(first.stdout, &candidate); err != nil {
		return materialized, registrationNativeFailure(err, "decode-candidate")
	}
	if err := compiled.ValidateCandidate(os.DirFS(workspace), candidate); err != nil {
		return materialized, registrationNativeFailure(err, "validate-candidate")
	}
	if !registrationContractMatches(candidate, action) {
		return materialized, newOperationError("OPERATION_INPUT", "bind-native-contract",
			"REGISTRATION_CONTRACT_SUBSTITUTED", "KNOWN_CONTRADICTION", "restore-compiled-native-binding")
	}
	verifier, failure := registrationVerifyCandidate(workspace, temporary, request, compiled, candidate)
	materialized.Verifier = verifier
	if failure != nil {
		return materialized, failure
	}
	materialized, failure = finishRegistrationMaterialization(plan, action, candidate, materialized)
	if failure != nil {
		return materialized, failure
	}
	return bindRegistrationInputEvidence(materialized, inputBinding)
}

func registrationNativeFailure(err error, step string) *operationError {
	if observed, ok := errors.AsType[*syntaxregistration.Failure](err); ok {
		class := observed.UnknownClass
		if observed.State == "REFUTED" {
			class = "KNOWN_CONTRADICTION"
		}
		failure := newOperationError(observed.Stage, observed.Step, observed.Reason, class, observed.NextOperation)
		failure.blockedBy = append([]string{}, observed.BlockedBy...)
		return failure
	}
	failure := newOperationError("SYNTAX_REGISTRATION", step, "REGISTRATION_INPUT_UNAVAILABLE",
		"DIRECT_MISSING", "restore-exact-registration-input")
	failure.diagnostics = []string{err.Error()}
	return failure
}

func registrationWorkerFailure(process registrationProcess) *operationError {
	first, _, _ := bytes.Cut(process.stderr, []byte("\n"))
	var observed syntaxregistration.Failure
	if json.Unmarshal(first, &observed) == nil && observed.Stage != "" &&
		(observed.State == "REFUTED" || observed.State == "UNKNOWN") {
		return registrationNativeFailure(&observed, "execute-worker")
	}
	return registrationProcessFailure("execute-worker", process)
}
