package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

var revisionConsumerPath = flag.String("revision-consumer", "", "compose the Gooo revision with an explicit trusted read-only consumer executable")
var revisionConsumerDigest = flag.String("revision-consumer-digest", "", "required sha256 identity of the consumer executable")

type revisionIndependentObservation struct {
	Schema string `json:"schema"`
	Decision string `json:"decision"`
	Reason string `json:"reason"`
	Operation *policycompilation.PolicyRevisionOperationObservation `json:"operation,omitempty"`
	Consumer revisionConsumerProcess `json:"consumer_process"`
	Pending *policycompilation.PolicyRevisionPending `json:"pending,omitempty"`
	Improvement string `json:"improvement"`
	ExecutionBoundary string `json:"execution_boundary"`
	MutationAuthority int `json:"mutation_authority"`
	PromotionAuthority int `json:"promotion_authority"`
}

func validateRevisionConsumerFlags(flags *flag.FlagSet) error {
	for _, name := range []string{"revision-consumer", "revision-consumer-digest", "revision-operation"} {
		value := flags.Lookup(name)
		if value == nil || value.Value.String() == "" {
			return fmt.Errorf("independent revision composition requires -%s", name)
		}
	}
	if !policycompilation.ValidDigest(flags.Lookup("revision-consumer-digest").Value.String()) {
		return errors.New("independent revision composition requires a sha256 consumer digest")
	}
	return nil
}

func revisionCompositionUnknown(report *revisionIndependentObservation, stage, reason, next string) {
	report.Decision, report.Reason = "UNKNOWN", reason
	report.Pending = &policycompilation.PolicyRevisionPending{
		State: "UNKNOWN", Stage: stage, Step: "COMPOSE_INDEPENDENT_REVISION_OBSERVATION",
		Reason: reason, UnknownClass: "DIRECT_MISSING", NextOperation: next, BlockedBy: []string{},
	}
}

func observeIndependentRevision(ctx context.Context, filename string, source, request []byte,
	pkg, namespace, operationPath, consumerPath, consumerDigest string, output io.Writer) error {
	report := revisionIndependentObservation{
		Schema: "gooo/meta-policy-revision-independent-observation/v1", Improvement: "UNKNOWN",
		ExecutionBoundary: "CALLER_PINNED_CONSUMER_NOT_AN_OS_SANDBOX",
		Consumer: revisionConsumerProcess{ExpectedExecutableDigest: consumerDigest, ExitCode: -1},
	}
	revisionCompositionUnknown(&report, "EXECUTABLE_BINDING", "CONSUMER_NOT_OBSERVED", "SUPPLY_PINNED_CONSUMER")
	observationError := composeIndependentRevision(ctx, filename, source, request, pkg, namespace,
		operationPath, consumerPath, consumerDigest, &report)
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return errors.Join(observationError, encoder.Encode(report))
}

func composeIndependentRevision(ctx context.Context, filename string, source, request []byte,
	pkg, namespace, operationPath, consumerPath, consumerDigest string, report *revisionIndependentObservation) error {
	directory, err := os.MkdirTemp("", "gooo-independent-revision-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	executable, observedDigest, err := pinRevisionConsumer(directory, consumerPath, consumerDigest)
	report.Consumer.ObservedExecutableDigest = observedDigest
	if err != nil {
		revisionCompositionUnknown(report, "EXECUTABLE_BINDING", "CONSUMER_EXECUTABLE_UNAVAILABLE_OR_MISMATCHED", "SUPPLY_PINNED_CONSUMER")
		return err
	}
	contract, err := os.ReadFile(operationPath)
	if err != nil {
		revisionCompositionUnknown(report, "GOOO_BINDING", "REVISION_OPERATION_UNAVAILABLE", "SUPPLY_GOOO_REVISION_OPERATION")
		return err
	}
	operation, err := policycompilation.ObserveGoooPolicyDecisionRevision(ctx, contract, filename, source, pkg, namespace, request)
	report.Operation = &operation
	if err != nil || operation.Observation == nil {
		revisionCompositionUnknown(report, "REVISION_EXECUTION", "REVISION_EXECUTION_INCOMPLETE", "OBSERVE_SOURCE_BOUND_REVISION")
		return errors.Join(err, errors.New("source-bound revision observation did not complete"))
	}
	payload, err := json.Marshal(operation.Observation)
	if err != nil {
		return err
	}
	process, runError := runRevisionConsumer(ctx, directory, executable, source, request, payload, pkg, namespace)
	process.ExpectedExecutableDigest, process.ObservedExecutableDigest = consumerDigest, observedDigest
	report.Consumer = process
	decision, validationError := checkRevisionConsumerReport([]byte(process.Stdout), operation, payload)
	if decision == "REFUTED" {
		report.Decision, report.Reason, report.Pending = "REFUTED", "INDEPENDENT_RECONSTRUCTION_REFUTED", nil
		return errors.Join(runError, validationError)
	}
	if runError != nil {
		revisionCompositionUnknown(report, "CONSUMER_PROCESS", "CONSUMER_EXECUTION_INCOMPLETE", "RERUN_PINNED_CONSUMER")
		return runError
	}
	if validationError != nil {
		revisionCompositionUnknown(report, "CONSUMER_REPORT", "CONSUMER_REPORT_NOT_ADMISSIBLE", "SUPPLY_BOUND_CONSUMER_REPORT")
		return validationError
	}
	report.Decision, report.Reason, report.Pending = decision, "FRESH_REPORT_INDEPENDENTLY_RECONSTRUCTED", nil
	return nil
}
