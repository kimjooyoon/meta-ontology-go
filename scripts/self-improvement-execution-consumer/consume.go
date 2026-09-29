package main

import (
	"errors"
	"fmt"
	"time"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func consume(g grant.LiveReport, c contract.LiveReport, m metadata, reserved reservation, outputPath string) error {
	started := time.Now()
	input := c.ContractResolution.ExecutionInput
	compileStarted := time.Now()
	plan, err := valueexecution.CompilePlan(input.Source.Path, []byte(input.Source.Bytes))
	compileElapsed := time.Since(compileStarted).Nanoseconds()
	if err != nil {
		return fmt.Errorf("compile exact execution input: %w", err)
	}
	if plan.SourceDigest != input.Source.Digest {
		return errors.New("compiled plan source digest does not match the authorized snapshot")
	}
	report := consumptionReport{Schema: reportSchema, GrantID: g.Resolution.Receipt.GrantID,
		GrantRequestDigest: g.Request.Digest, GrantReceiptDigest: g.Resolution.Receipt.Digest,
		ReservationDigest: reserved.Digest, SubjectSHA: m.SubjectSHA,
		GrantRunID: m.GrantRunID, GrantRunAttempt: m.GrantAttempt,
		GrantArtifactID: m.GrantArtifactID, GrantArtifactDigest: m.GrantArtifactDigest,
		ContractRunID: m.ContractRunID, ContractRunAttempt: m.ContractAttempt,
		ContractArtifactID: m.ContractArtifactID, ContractArtifactDigest: m.ContractArtifactDigest,
		CandidateStableID: input.CandidateStableID, CandidateDigest: input.CandidateDigest,
		ExecutionInputDigest: input.Digest, ObservationDigest: input.ObservationDigest,
		OperationID:    input.OperationID,
		ExecutionCount: 1, EvaluatorInvocations: len(input.Corpus), ConsumedUses: 1,
		RemainingUses: 0, OneUseEnforced: true, RepositoryWrites: 0, ExternalEffects: []string{},
		PlanSourceDigest: plan.SourceDigest, PlanIdentityDigest: planIdentityDigest(plan),
		CompileElapsedNanoseconds: compileElapsed, PerformanceImprovement: "UNKNOWN",
		CaseCount: len(input.Corpus), Cases: make([]caseResult, 0, len(input.Corpus))}
	evaluationStarted := time.Now()
	for _, test := range input.Corpus {
		caseStarted := time.Now()
		execution, runErr := plan.Execute(map[string]int64{input.Activity.Name: test.Input})
		result := caseResult{ID: test.ID, Input: test.Input, ExpectedOutput: test.ExpectedOutput,
			ExecutionDigest: execution.ExecutionDigest}
		if runErr != nil {
			result.Error = runErr.Error()
		}
		actual, ok := execution.Results[input.Activity.Name]
		if ok {
			value := actual.Value
			result.ActualOutput, result.ResultDigest = &value, actual.ResultDigest
		}
		result.ElapsedNanoseconds = time.Since(caseStarted).Nanoseconds()
		result.Passed = runErr == nil && ok && execution.Phase == valueexecution.ExecutionPhaseCompleted &&
			execution.ApplyCalls == 1 &&
			actual.Value == test.ExpectedOutput && actual.SourceDigest == input.Source.Digest &&
			actual.SemanticFingerprint == input.Activity.SemanticFingerprint &&
			actual.OutputEntity == input.Activity.OutputEntity && actual.ProducerActivity == input.Activity.Name &&
			validDigest(actual.ResultDigest)
		if result.Passed {
			report.PassedCases++
		}
		report.Cases = append(report.Cases, result)
	}
	report.EvaluationElapsedNanoseconds = time.Since(evaluationStarted).Nanoseconds()
	report.ElapsedNanoseconds = time.Since(started).Nanoseconds()
	report.Digest = consumptionDigest(report)
	if err := writeJSON(outputPath, report); err != nil {
		return err
	}
	if report.PassedCases != report.CaseCount {
		return errors.New("bounded execution produced failing or incomplete value cases")
	}
	return nil
}

func planIdentityDigest(plan valueexecution.Plan) string {
	return digestJSON(struct {
		SourceDigest        string `json:"source_digest"`
		SemanticFingerprint string `json:"semantic_fingerprint"`
	}{plan.SourceDigest, plan.SemanticFingerprint})
}

func consumptionDigest(report consumptionReport) string {
	report.Digest = ""
	return digestJSON(report)
}

func reservationDigest(value reservation) string {
	value.Digest = ""
	return digestJSON(value)
}

func validateReservation(value reservation) error {
	if value.Schema != reportSchema+"/reservation" || value.Status != "CONSUMPTION_RESERVED" ||
		!value.OneUseEnforced || value.GrantID == "" || value.SourceRunID <= 0 ||
		value.SourceAttempt != 1 || value.ExecutionCount != 0 || value.ConsumedUses != 1 ||
		value.RemainingUses != 0 || value.Digest != reservationDigest(value) {
		return errors.New("consumption reservation is invalid")
	}
	return nil
}
