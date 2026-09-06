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
	Variant      string                     `json:"variant"`
	Command      []string                   `json:"command"`
	ExitCode     int                        `json:"exit_code"`
	WallMS       int64                      `json:"wall_ms"`
	StdoutDigest string                     `json:"stdout_digest"`
	StderrDigest string                     `json:"stderr_digest"`
	Events       []CallbackPackageTestEvent `json:"events"`
}

type CallbackExtractionObservation struct {
	Schema             string                              `json:"schema"`
	Scope              string                              `json:"scope"`
	Decision           string                              `json:"decision"`
	SourceDigest       string                              `json:"source_digest"`
	SourcePackageDigest string                             `json:"source_package_digest"`
	FinalPackageDigest string                              `json:"final_package_digest"`
	ProposalDigest     string                              `json:"proposal_digest"`
	ContractDigest     string                              `json:"contract_digest"`
	ModulePath         string                              `json:"module_path"`
	GoVersion          string                              `json:"go_version"`
	GeneratedFiles     int                                 `json:"generated_files"`
	DependencyBinding  string                              `json:"dependency_binding"`
	Runs               []CallbackPackageRun                `json:"runs"`
	TestEventDigest    string                              `json:"test_event_digest"`
	Record             generation.CallbackExtractionRecord `json:"record"`
	Frontier           CallbackExtractionClaim             `json:"frontier"`
	OperationAdmission string                              `json:"operation_admission"`
	ApplyPermission    string                              `json:"apply_permission"`
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
	if ctx == nil {
		return observation, fmt.Errorf("callback package observation requires a bounded context")
	}
	if _, bounded := ctx.Deadline(); !bounded {
		return observation, fmt.Errorf("callback package observation requires a deadline")
	}
	proposal, err := PlanCallbackExtraction(root, logical, subject)
	if err != nil {
		return observation, err
	}
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
	baseline, final, err := callbackObservationSources(root, logical, proposal)
	if err != nil {
		return observation, err
	}
	observation.SourcePackageDigest = proofDigest(generatedPackagePayload(baseline))
	observation.FinalPackageDigest = proofDigest(generatedPackagePayload(final))
	observation.ModulePath, observation.GoVersion, err = callbackObservationToolchain(ctx, root)
	if err != nil {
		return observation, err
	}
	directory, err := os.MkdirTemp("", "gooo-callback-observation-")
	if err != nil {
		return observation, err
	}
	defer os.RemoveAll(directory)
	for _, variant := range []struct {
		name  string
		files map[string][]byte
	}{{"source", baseline}, {"final", final}} {
		workdir, prepareError := materializeCallbackObservation(directory, variant.name, root, logical, observation.ModulePath, observation.GoVersion, variant.files)
		if prepareError != nil {
			return observation, prepareError
		}
		run, runError := runCallbackPackageObservation(ctx, workdir, variant.name, requiredTest)
		observation.Runs = append(observation.Runs, run)
		if runError != nil {
			observation.Frontier.Step = "RUN_" + strings.ToUpper(variant.name) + "_PACKAGE"
			observation.Frontier.Reason = "PACKAGE_TEST_OBSERVATION_INCOMPLETE"
			observation.Frontier.NextOperation = "RESOLVE_PACKAGE_TEST_OBSERVATION"
			if run.ExitCode > 0 {
				observation.Decision = "REFUTED"
				observation.Frontier.State = "REFUTED"
				observation.Frontier.Reason = "PACKAGE_TESTS_FAILED"
				observation.Frontier.UnknownClass = ""
				observation.Frontier.NextOperation = "PRESERVE_PACKAGE_TEST_COUNTEREXAMPLE"
			}
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
	raw, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	record, err := contract.BuildRecord(4, "UNKNOWN", proofDigest(raw), len(observation.Runs), 2)
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
	args := []string{"test", "-mod=mod", "-json", "-count=1", "."}
	run := CallbackPackageRun{Variant: variant, Command: append([]string{"go"}, args...), ExitCode: -1}
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	started := time.Now()
	err := command.Run()
	run.WallMS = time.Since(started).Milliseconds()
	run.StdoutDigest, run.StderrDigest = proofDigest(stdout.Bytes()), proofDigest(stderr.Bytes())
	if command.ProcessState != nil {
		run.ExitCode = command.ProcessState.ExitCode()
	}
	if err != nil {
		return run, fmt.Errorf("%s package tests: %w; stderr=%s", variant, err, stderr.String())
	}
	run.Events, err = callbackPackageTestEvents(stdout.Bytes(), requiredTest)
	return run, err
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
		return nil, fmt.Errorf("required source test %s did not pass", requiredTest)
	}
	slices.SortFunc(events, func(left, right CallbackPackageTestEvent) int { return strings.Compare(left.Name, right.Name) })
	return events, nil
}
