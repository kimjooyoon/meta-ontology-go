package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

type upstreamObservation struct {
	ID                  string                                `json:"id"`
	InputAuthority      string                                `json:"input_authority"`
	RequestedOperation  string                                `json:"requested_operation"`
	DeclaredCohort      []metricstrategy.ConceptOperationSpec `json:"declared_cohort"`
	ExpectedCount       int                                   `json:"expected_count"`
	AddedExpectedCount  int                                   `json:"added_expected_count"`
	InputDigests        map[string]string                     `json:"input_digests"`
	Refusal             *unresolvedpolicy.OperationRefusal    `json:"refusal,omitempty"`
	Consumer            *programverify.RefusalReport          `json:"consumer,omitempty"`
	CoverageBPS         *int                                  `json:"coverage_bps"`
	ProgramProduced     bool                                  `json:"program_produced"`
	ConsumerAttempted   bool                                  `json:"consumer_attempted"`
	ReadinessCredit     bool                                  `json:"readiness_credit"`
	RepositoryWrites    int                                   `json:"repository_writes"`
	PromotionAuthorized bool                                  `json:"promotion_authorized"`
	ElapsedNanoseconds  int64                                 `json:"elapsed_nanoseconds"`
	Passed              bool                                  `json:"passed"`
	Error               string                                `json:"error,omitempty"`
}

func observeUnregisteredUpstream(cfg config, f fixture) (result upstreamObservation, observationErr error) {
	started := time.Now()
	defer func() {
		result.ElapsedNanoseconds = time.Since(started).Nanoseconds()
		if observationErr != nil {
			result.Error = observationErr.Error()
		}
	}()
	result.ID, result.RequestedOperation = "unregistered-upstream-operation", unregisteredUpstreamOperation
	result.InputAuthority = "ADVERSARIAL_RESEALED_LEDGER_NOT_UPSTREAM_VERIFICATION_AUTHORITY"
	inputs, cohort, err := extendUnregisteredLedger(f.Inputs)
	if err != nil {
		return result, err
	}
	result.DeclaredCohort, result.ExpectedCount = cohort, len(cohort)
	result.AddedExpectedCount = len(cohort) - len(f.Contract)
	inputs, err = synthesizeUnregisteredStrategy(cfg, inputs)
	if err != nil {
		return result, err
	}
	program, source, compileErr := metricprogram.Compile(inputs.Strategy, inputs.StrategyVerification)
	result.ProgramProduced = program.Schema != "" || source != nil
	if !errors.As(compileErr, &result.Refusal) {
		return result, fmt.Errorf("upstream operation lost typed source origin: %v", compileErr)
	}
	raw, err := json.Marshal(result.Refusal)
	if err != nil {
		return result, err
	}
	result.ConsumerAttempted = true
	consumer, err := programverify.VerifyOperationRefusal(os.DirFS(cfg.Root), cfg.Origin.Repository,
		cfg.Origin.SubjectSHA, inputs.Strategy, inputs.StrategyVerification, raw)
	if err != nil {
		return result, err
	}
	result.Consumer = &consumer
	result.InputDigests = inputDigests(variant{Inputs: inputs})
	result.InputDigests["operation-refusal.json"] = artifact.ContentDigest(raw)
	result.Passed = validUpstreamObservation(result, f)
	if err := saveUpstreamObservation(cfg, f, inputs, result, raw); err != nil {
		return result, err
	}
	if !result.Passed {
		return result, fmt.Errorf("unregistered upstream origin or authority boundary diverged")
	}
	return result, nil
}

func validUpstreamObservation(result upstreamObservation, f fixture) bool {
	refusal, consumer := result.Refusal, result.Consumer
	if refusal == nil || consumer == nil || result.ProgramProduced || result.AddedExpectedCount != 1 ||
		result.ExpectedCount != len(result.DeclaredCohort) || result.ExpectedCount != len(f.Contract)+1 ||
		result.ReadinessCredit || result.RepositoryWrites != 0 || result.PromotionAuthorized || result.CoverageBPS != nil {
		return false
	}
	if refusal.Origin.RequestedOperation != result.RequestedOperation || !refusal.OriginRecorded ||
		refusal.Reason != "OPERATION_NOT_IN_SOURCE_REGISTRY" || refusal.Resolution != "invariant_only" ||
		refusal.NextOperation != "DECLARE_SOURCE_OPERATION" || refusal.ReadinessCredit ||
		refusal.RepositoryWorkspaceWrites || refusal.PromotionAuthorized ||
		consumer.Origin != refusal.Origin || consumer.Status != "VERIFIED_SOURCE_ORIGIN_REFUSAL" ||
		consumer.UpstreamVerification != "UNASSESSED" || consumer.ReadinessCredit ||
		consumer.RepositoryWorkspaceWrites || consumer.PromotionAuthorized {
		return false
	}
	bySubject := make(map[string]metricstrategy.ConceptOperationSpec)
	for _, member := range result.DeclaredCohort {
		if _, duplicate := bySubject[member.Subject]; duplicate {
			return false
		}
		bySubject[member.Subject] = member
	}
	for _, original := range f.Contract {
		if bySubject[original.Subject] != original {
			return false
		}
	}
	added, ok := bySubject[result.RequestedOperation]
	return ok && added.Carrier == result.RequestedOperation && added.Family == "FOUNDATION" && added.Trilemma == "AXIOM"
}

func saveUpstreamObservation(cfg config, f fixture, inputs co.SourceInputs, result upstreamObservation, refusal []byte) error {
	directory := filepath.Join(cfg.Output, result.ID)
	if err := saveVariant(directory, cfg.Origin, f, variant{Inputs: inputs}); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "operation-refusal.json"), refusal, 0640); err != nil {
		return err
	}
	consumer, err := json.MarshalIndent(result.Consumer, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "refusal-verification.json"), append(consumer, '\n'), 0640)
}
