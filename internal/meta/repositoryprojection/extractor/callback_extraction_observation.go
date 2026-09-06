package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type CallbackPackageTestEvent struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

type CallbackPackageRun struct {
	Variant            string                     `json:"variant"`
	Command            []string                   `json:"command"`
	ExitCode           int                        `json:"exit_code"`
	WallMS             int64                      `json:"wall_ms"`
	StdoutDigest       string                     `json:"stdout_digest"`
	StderrDigest       string                     `json:"stderr_digest"`
	Stdout             []byte                     `json:"stdout"`
	Stderr             []byte                     `json:"stderr"`
	Events             []CallbackPackageTestEvent `json:"events"`
	TestEventsComplete bool                       `json:"test_events_complete"`
}

type CallbackExtractionObservation struct {
	Schema              string                              `json:"schema"`
	Scope               string                              `json:"scope"`
	Decision            string                              `json:"decision"`
	SourceDigest        string                              `json:"source_digest"`
	SourcePackageDigest string                              `json:"source_package_digest"`
	FinalPackageDigest  string                              `json:"final_package_digest"`
	ProposalDigest      string                              `json:"proposal_digest"`
	ContractDigest      string                              `json:"contract_digest"`
	ModulePath          string                              `json:"module_path"`
	ModuleDigest        string                              `json:"module_snapshot_digest"`
	ModuleFiles         int                                 `json:"module_snapshot_files"`
	ModuleBytes         int64                               `json:"module_snapshot_bytes"`
	GoVersion           string                              `json:"go_version"`
	GeneratedFiles      int                                 `json:"generated_files"`
	AttemptedRuns       int                                 `json:"attempted_runs"`
	CompletedTestRuns   int                                 `json:"completed_test_runs"`
	RequiredTestRuns    int                                 `json:"required_test_runs"`
	DependencyBinding   string                              `json:"dependency_binding"`
	Runs                []CallbackPackageRun                `json:"runs"`
	TestEventDigest     string                              `json:"test_event_digest"`
	Record              generation.CallbackExtractionRecord `json:"record"`
	Frontier            CallbackExtractionClaim             `json:"frontier"`
	OperationAdmission  string                              `json:"operation_admission"`
	ApplyPermission     string                              `json:"apply_permission"`
}

// ObserveCallbackExtraction executes only in CI and only in disposable copies.
// Test-event agreement does not grant semantic admission or repository writes.
func ObserveCallbackExtraction(ctx context.Context, root, logical, subject string) (observation CallbackExtractionObservation, err error) {
	observation = CallbackExtractionObservation{
		Schema: "gooo/callback-extraction-observation/v1", Scope: "PACKAGE_TEST_EVENTS_ONLY", Decision: "UNKNOWN",
		DependencyBinding: "UNBOUND", OperationAdmission: "UNKNOWN", ApplyPermission: "FORBIDDEN",
		Frontier: CallbackExtractionClaim{State: "UNKNOWN", Stage: "CALLBACK_OBSERVATION", Step: "CHECK_EXECUTION_BOUNDARY",
			Reason: "CI_EXECUTION_REQUIRED", UnknownClass: "DIRECT_MISSING", NextOperation: "RUN_PACKAGE_OBSERVATION_IN_CI", BlockedBy: []string{}},
	}
	if os.Getenv("CI") != "true" {
		return observation, fmt.Errorf("callback package observation is CI-only")
	}
	observation.Frontier = callbackObservationUnknown("CHECK_DEADLINE", "BOUNDED_CONTEXT_REQUIRED", "UNBOUNDED", "SUPPLY_BOUNDED_OBSERVER_CONTEXT")
	if ctx == nil {
		return observation, fmt.Errorf("callback package observation requires a bounded context")
	}
	if _, bounded := ctx.Deadline(); !bounded {
		return observation, fmt.Errorf("callback package observation requires a deadline")
	}
	observation.Frontier = callbackObservationUnknown("PLAN_FINAL_PACKAGE", "CALLBACK_EXTRACTION_PROPOSAL_UNAVAILABLE", "DIRECT_MISSING", "PLAN_CALLBACK_EXTRACTION")
	proposal, err := PlanCallbackExtraction(root, logical, subject)
	if err != nil {
		return observation, err
	}
	observation.Frontier = callbackObservationUnknown("BIND_REQUIRED_TEST", "SOURCE_TEST_SUBJECT_UNAVAILABLE", "AMBIGUOUS", "SELECT_SOURCE_TEST_SUBJECT")
	requiredTest, ok := strings.CutPrefix(subject, "func:Test")
	if !ok || requiredTest == "" {
		return observation, fmt.Errorf("callback package observation requires a test function subject")
	}
	requiredTest = "Test" + requiredTest
	observation.SourceDigest = proposal.SourceDigest
	observation.ProposalDigest = proposal.PackageDigest
	observation.ContractDigest = proposal.Contract.SemanticDigest
	observation.GeneratedFiles = len(proposal.Artifacts)
	defer func() {
		recordError := bindCallbackPackageObservation(&observation, proposal.Contract)
		if err == nil {
			err = recordError
		}
	}()
	observation.Frontier = callbackObservationUnknown("SNAPSHOT_PACKAGE", "SOURCE_PACKAGE_SNAPSHOT_UNAVAILABLE", "DIRECT_MISSING", "CAPTURE_SOURCE_PACKAGE_SNAPSHOT")
	baseline, final, err := callbackObservationSources(root, logical, proposal)
	if err != nil {
		return observation, err
	}
	observation.SourcePackageDigest = proofDigest(generatedPackagePayload(baseline))
	observation.FinalPackageDigest = proofDigest(generatedPackagePayload(final))
	observation.Frontier = callbackObservationUnknown("SNAPSHOT_MODULE", "SOURCE_MODULE_SNAPSHOT_UNAVAILABLE", "DIRECT_MISSING", "CAPTURE_SOURCE_MODULE_SNAPSHOT")
	moduleFiles, err := callbackObservationModuleSources(ctx, root, logical, baseline)
	if err != nil {
		return observation, err
	}
	modulePayload, err := json.Marshal(moduleFiles)
	if err != nil {
		return observation, err
	}
	observation.ModuleDigest, observation.ModuleFiles = proofDigest(modulePayload), len(moduleFiles)
	for _, raw := range moduleFiles {
		observation.ModuleBytes += int64(len(raw))
	}
	observation.Frontier = callbackObservationUnknown("BIND_TOOLCHAIN", "GO_TOOLCHAIN_IDENTITY_UNAVAILABLE", "DIRECT_MISSING", "RESTORE_GO_TOOLCHAIN_OBSERVATION")
	observation.ModulePath, observation.GoVersion, err = callbackObservationToolchain(ctx, root)
	if err != nil {
		return observation, err
	}
	observation.Frontier = callbackObservationUnknown("CREATE_TEMP_WORKSPACE", "TEMPORARY_OUTPUT_UNAVAILABLE", "DIRECT_MISSING", "ALLOCATE_EXTERNAL_OBSERVATION_WORKSPACE")
	directory, err := os.MkdirTemp("", "gooo-callback-observation-")
	if err != nil {
		return observation, err
	}
	defer os.RemoveAll(directory)
	for _, variant := range []struct {
		name  string
		files map[string][]byte
	}{{"source", baseline}, {"final", final}} {
		observation.Frontier = callbackObservationUnknown("PREPARE_"+strings.ToUpper(variant.name)+"_PACKAGE", "PACKAGE_WORKSPACE_UNAVAILABLE", "DIRECT_MISSING", "MATERIALIZE_PACKAGE_OBSERVATION")
		workdir, prepareError := materializeCallbackObservation(directory, variant.name, logical, moduleFiles, variant.files)
		if prepareError != nil {
			return observation, prepareError
		}
		run, runError := runCallbackPackageObservation(ctx, workdir, variant.name, requiredTest)
		observation.Runs = append(observation.Runs, run)
		if runError != nil {
			observation.Frontier = callbackPackageFailureFrontier(run)
			observation.Decision = observation.Frontier.State
			return observation, runError
		}
	}
	if !slices.Equal(observation.Runs[0].Events, observation.Runs[1].Events) {
		observation.Decision = "REFUTED"
		observation.Frontier = CallbackExtractionClaim{State: "REFUTED", Stage: "CALLBACK_OBSERVATION", Step: "COMPARE_TEST_EVENTS",
			Reason: "PACKAGE_TEST_EVENT_MISMATCH", NextOperation: "PRESERVE_PACKAGE_TEST_COUNTEREXAMPLE", BlockedBy: []string{}}
		return observation, fmt.Errorf("source and final package test events differ")
	}
	events, err := json.Marshal(observation.Runs[0].Events)
	if err != nil {
		return observation, err
	}
	observation.TestEventDigest = proofDigest(events)
	observation.Decision = "OBSERVED"
	observation.Frontier = CallbackExtractionClaim{State: "UNKNOWN", Stage: "CALLBACK_OBSERVATION", Step: "SEMANTIC_ADMISSION",
		Reason: "FINITE_TEST_EVENTS_DO_NOT_PROVE_SEMANTIC_EQUIVALENCE", UnknownClass: "UNBOUNDED",
		NextOperation: "DECLARE_AND_PROVE_CALLBACK_OBSERVATION_CONTRACT", BlockedBy: []string{}}
	return observation, nil
}

func bindCallbackPackageObservation(observation *CallbackExtractionObservation, contract generation.CallbackExtractionContract) error {
	observation.AttemptedRuns, observation.CompletedTestRuns, observation.RequiredTestRuns = len(observation.Runs), 0, 2
	for _, run := range observation.Runs {
		if run.ExitCode == 0 && run.TestEventsComplete && len(run.Events) > 0 {
			observation.CompletedTestRuns++
		}
	}
	raw, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	record, err := contract.BuildRecord(4, "UNKNOWN", proofDigest(raw), observation.CompletedTestRuns, observation.RequiredTestRuns)
	if err != nil {
		return err
	}
	record.Fields["Scope"] = observation.Scope
	record.Fields["SourcePackageDigest"] = observation.SourcePackageDigest
	record.Fields["FinalPackageDigest"] = observation.FinalPackageDigest
	record.Fields["TestEventDigest"] = observation.TestEventDigest
	record.Fields["ObservationDecision"] = observation.Decision
	observation.Record = record
	return nil
}

func runCallbackPackageObservation(ctx context.Context, directory, variant, requiredTest string) (CallbackPackageRun, error) {
	args := []string{"test", "-mod=readonly", "-json", "-count=1", "."}
	run := CallbackPackageRun{Variant: variant, Command: append([]string{"go"}, args...), ExitCode: -1}
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	started := time.Now()
	err := command.Run()
	run.WallMS = time.Since(started).Milliseconds()
	run.Stdout, run.Stderr = bytes.Clone(stdout.Bytes()), bytes.Clone(stderr.Bytes())
	run.StdoutDigest, run.StderrDigest = proofDigest(run.Stdout), proofDigest(run.Stderr)
	if command.ProcessState != nil {
		run.ExitCode = command.ProcessState.ExitCode()
	}
	events, eventError := callbackPackageTestEvents(run.Stdout, requiredTest)
	run.Events = events
	run.TestEventsComplete = err == nil && eventError == nil
	if err != nil {
		return run, fmt.Errorf("%s package command: %w; stdout=%s; stderr=%s", variant, err, run.Stdout, run.Stderr)
	}
	return run, eventError
}

func callbackObservationUnknown(step, reason, class, next string) CallbackExtractionClaim {
	return CallbackExtractionClaim{ID: "gooo://callback-extraction/claim/observers", State: "UNKNOWN",
		Stage: "CALLBACK_OBSERVATION", Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: []string{}}
}

func callbackPackageFailureFrontier(run CallbackPackageRun) CallbackExtractionClaim {
	claim := callbackObservationUnknown("RUN_"+strings.ToUpper(run.Variant)+"_PACKAGE",
		"PACKAGE_TEST_OBSERVATION_INCOMPLETE", "DIRECT_MISSING", "RESOLVE_PACKAGE_TEST_OBSERVATION")
	for _, event := range run.Events {
		if event.Action == "fail" {
			claim.State, claim.Reason, claim.UnknownClass = "REFUTED", "PACKAGE_TEST_COUNTEREXAMPLE_OBSERVED", ""
			claim.NextOperation = "PRESERVE_PACKAGE_TEST_COUNTEREXAMPLE"
		}
	}
	return claim
}

func callbackPackageTestEvents(raw []byte, requiredTest string) ([]CallbackPackageTestEvent, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	events := []CallbackPackageTestEvent{}
	seen, requiredPassed := map[string]bool{}, false
	for {
		var event struct{ Action, Test string }
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("package test event decoding: %w", err)
		}
		if event.Test == "" {
			continue
		}
		switch event.Action {
		case "run", "output", "pause", "cont", "bench":
			continue
		case "pass", "fail", "skip":
		default:
			return nil, fmt.Errorf("unknown package test action %q", event.Action)
		}
		if seen[event.Test] {
			return nil, fmt.Errorf("duplicate terminal test event for %s", event.Test)
		}
		seen[event.Test] = true
		events = append(events, CallbackPackageTestEvent{Name: event.Test, Action: event.Action})
		requiredPassed = requiredPassed || event.Test == requiredTest && event.Action == "pass"
	}
	if !requiredPassed {
		return events, fmt.Errorf("required source test %s did not pass", requiredTest)
	}
	slices.SortFunc(events, func(left, right CallbackPackageTestEvent) int { return strings.Compare(left.Name, right.Name) })
	return events, nil
}
