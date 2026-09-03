package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const adoptionUsage = "usage: gooo adopt <observation.gooo> --input <file.gooo> --observation FILE --proposal FILE --authorization FILE --out <directory>"

type adoptionOptions struct {
	contractFilename      string
	inputFilename         string
	observationFilename   string
	proposalFilename      string
	authorizationFilename string
	outputDir             string
}

type adoptionPair struct {
	first                generationResult
	replay               generationResult
	firstSemanticDigest  cache.Digest
	replaySemanticDigest cache.Digest
	operationCount       int
	cacheMisses          int
	cacheHits            int
	metrics              generation.SemanticAdoptionRuntimeMetrics
}

func runAdoption(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseAdoptionArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	inputs, diagnostics, err := loadObservationInputs(observeOptions{contractFilename: options.contractFilename, inputFilename: options.inputFilename}, reader, parser)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if diagnostics.HasErrors() {
		_ = reportDiagnostics(diagnostics, stderr)
		return exitFailure
	}
	if !reportDiagnostics(diagnostics, stderr) {
		return exitFailure
	}
	report, err := buildAdoptionReport(options, inputs, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: adoption: %v\n", err)
		return exitFailure
	}
	return writeAdoptionReport(options.outputDir, report, stdout, stderr)
}

func buildAdoptionReport(options adoptionOptions, inputs observationInputs, reader SourceReader) (generation.SemanticAdoptionReport, error) {
	observationData, observation, proposalData, proposal, authorizationData, authorization, err := readAdoptionInputs(options, reader)
	if err != nil {
		return generation.SemanticAdoptionReport{}, err
	}
	if err := validateAdoptionInputs(inputs, observationData, observation, proposalData, proposal, authorization); err != nil {
		return generation.SemanticAdoptionReport{}, err
	}
	proposalDigest := cache.HashBytes(proposalData).String()
	authorizationDigest := cache.HashBytes(authorizationData).String()
	evidence := generation.SemanticAdoptionEvidence{
		Schema: generation.SemanticAdoptionEvidenceSchema, ProposalDigest: proposalDigest,
		AuthorizationDigest: authorizationDigest, CandidateStableID: proposal.Candidate.StableID,
		InputDigest: proposal.Candidate.InputDigest, RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	var before, adopted adoptionPair
	if authorization.Authorized {
		before, err = runAdoptionPair(inputs.file, authorization, false)
		if err == nil {
			adopted, err = runAdoptionPair(inputs.file, authorization, true)
		}
		if err != nil {
			return generation.SemanticAdoptionReport{}, fmt.Errorf("compiler execution: %w", err)
		}
		evidence = adoptionEvidence(proposal, proposalDigest, authorizationDigest, before, adopted)
	} else {
		evidence.Decision = "UNKNOWN"
		evidence.Reason = generation.SemanticAdoptionUnknownReason
		evidence.Unknown = generation.AdoptionUnknownState()
	}
	decision, reason, unknown, err := generation.VerifySemanticAdoption(proposal, proposalDigest, authorization, authorizationDigest, evidence)
	if err != nil {
		return generation.SemanticAdoptionReport{}, fmt.Errorf("independent verification: %w", err)
	}
	evidence.Decision, evidence.Reason, evidence.Unknown = decision, reason, unknown
	boundObservation := observation
	if decision == "CLOSED" {
		boundObservation.PairEvidence = generation.SemanticObservationPairEvidence{
			EvidenceAvailable: true, ChangeAdopted: true, BehaviorEqual: evidence.BehaviorEqual,
			DeterminismEqual: evidence.DeterminismEqual, BeforeOperationCount: evidence.BeforeOperationCount,
			AfterOperationCount: evidence.AfterOperationCount,
		}
		boundObservation.Metrics.BeforeOperationCount = evidence.BeforeOperationCount
		boundObservation.Metrics.AfterOperationCount = evidence.AfterOperationCount
		boundObservation.Adoption = &evidence
	} else if decision == "UNKNOWN" {
		boundObservation.Decision = "UNKNOWN"
		boundObservation.Reason = generation.SemanticObservationUnknownReason
		boundObservation.Unknown = generation.SemanticObservationUnknownState()
		boundObservation.PairEvidence = generation.SemanticObservationPairEvidence{}
		boundObservation.Metrics.BeforeOperationCount = 0
		boundObservation.Metrics.AfterOperationCount = 0
	} else if decision == generation.SemanticAdoptionRefuted {
		boundObservation.Decision = generation.SemanticAdoptionRefuted
		boundObservation.Reason = generation.SemanticObservationContradiction
		boundObservation.Unknown = nil
		boundObservation.PairEvidence = generation.SemanticObservationPairEvidence{
			EvidenceAvailable: true, ChangeAdopted: true, BehaviorEqual: evidence.BehaviorEqual,
			DeterminismEqual: evidence.DeterminismEqual, BeforeOperationCount: evidence.BeforeOperationCount,
			AfterOperationCount: evidence.AfterOperationCount, Contradiction: generation.SemanticObservationContradiction,
		}
		boundObservation.Adoption = &evidence
	}
	return generation.SemanticAdoptionReport{
		Schema: generation.SemanticAdoptionReportSchema, Lifecycle: adoptionLifecycle(authorization.Authorized),
		ObservationDigest: cache.HashBytes(observationData).String(), ProposalDigest: proposalDigest,
		AuthorizationDigest: authorizationDigest, Proposal: proposal, Authorization: authorization,
		Evidence: evidence, Observation: boundObservation, BeforeRuntimeMetrics: before.metrics,
		AfterRuntimeMetrics: adopted.metrics, IndependentDecision: decision, IndependentReason: reason,
		RepositoryWrites: 0, LocalTestExecutions: 0,
	}, nil
}

func writeAdoptionReport(outputDir string, report generation.SemanticAdoptionReport, stdout, stderr io.Writer) int {
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: adoption: encode report: %v\n", err)
		return exitFailure
	}
	reportData = append(reportData, '\n')
	if err := writeAdoptionArtifact(outputDir, "adoption-result.json", reportData); err != nil {
		fmt.Fprintf(stderr, "gooo: adoption: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "adoption: %s (%s)\n", outputDir, report.IndependentDecision)
	return exitOK
}

func parseAdoptionArguments(args []string) (adoptionOptions, error) {
	if len(args) == 0 {
		return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
	}
	options := adoptionOptions{contractFilename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" {
			return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
		}
		value := args[index+1]
		switch args[index] {
		case "--input":
			if options.inputFilename != "" {
				return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
			}
			options.inputFilename = value
		case "--observation":
			if options.observationFilename != "" {
				return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
			}
			options.observationFilename = value
		case "--proposal":
			if options.proposalFilename != "" {
				return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
			}
			options.proposalFilename = value
		case "--authorization":
			if options.authorizationFilename != "" {
				return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
			}
			options.authorizationFilename = value
		case "--out":
			if options.outputDir != "" {
				return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
			}
			options.outputDir = value
		default:
			return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
		}
		index++
	}
	if options.contractFilename == "" || options.inputFilename == "" || options.observationFilename == "" ||
		options.proposalFilename == "" || options.authorizationFilename == "" || options.outputDir == "" {
		return adoptionOptions{}, fmt.Errorf("%s", adoptionUsage)
	}
	return options, nil
}

func readAdoptionInputs(options adoptionOptions, reader SourceReader) ([]byte, generation.SemanticObservation, []byte, generation.SemanticAdoptionProposal, []byte, generation.SemanticAdoptionAuthorization, error) {
	observationData, err := reader.ReadFile(options.observationFilename)
	if err != nil {
		return nil, generation.SemanticObservation{}, nil, generation.SemanticAdoptionProposal{}, nil, generation.SemanticAdoptionAuthorization{}, fmt.Errorf("read observation: %w", err)
	}
	proposalData, err := reader.ReadFile(options.proposalFilename)
	if err != nil {
		return nil, generation.SemanticObservation{}, nil, generation.SemanticAdoptionProposal{}, nil, generation.SemanticAdoptionAuthorization{}, fmt.Errorf("read proposal: %w", err)
	}
	authorizationData, err := reader.ReadFile(options.authorizationFilename)
	if err != nil {
		return nil, generation.SemanticObservation{}, nil, generation.SemanticAdoptionProposal{}, nil, generation.SemanticAdoptionAuthorization{}, fmt.Errorf("read authorization: %w", err)
	}
	var observation generation.SemanticObservation
	var proposal generation.SemanticAdoptionProposal
	var authorization generation.SemanticAdoptionAuthorization
	if err := json.Unmarshal(observationData, &observation); err != nil {
		return nil, observation, nil, proposal, nil, authorization, fmt.Errorf("decode observation: %w", err)
	}
	if err := json.Unmarshal(proposalData, &proposal); err != nil {
		return nil, observation, nil, proposal, nil, authorization, fmt.Errorf("decode proposal: %w", err)
	}
	if err := json.Unmarshal(authorizationData, &authorization); err != nil {
		return nil, observation, nil, proposal, nil, authorization, fmt.Errorf("decode authorization: %w", err)
	}
	return observationData, observation, proposalData, proposal, authorizationData, authorization, nil
}

func validateAdoptionInputs(inputs observationInputs, observationData []byte, observation generation.SemanticObservation, proposalData []byte, proposal generation.SemanticAdoptionProposal, authorization generation.SemanticAdoptionAuthorization) error {
	if err := generation.VerifySemanticObservation(observation); err != nil {
		return fmt.Errorf("observation verification: %w", err)
	}
	if observation.Decision != "CLOSED" || len(observation.Candidates) != 1 {
		return fmt.Errorf("observation must be CLOSED with one candidate")
	}
	if observation.ContractDigest != cache.HashBytes(inputs.contractSource).String() || observation.InputSourceDigest != cache.HashBytes(inputs.inputSource).String() {
		return fmt.Errorf("observation is not bound to the supplied contract and input")
	}
	if proposal.ObservationDigest != cache.HashBytes(observationData).String() || proposal.ContractDigest != observation.ContractDigest || proposal.InputSourceDigest != observation.InputSourceDigest {
		return fmt.Errorf("proposal is not bound to the exact observation inputs")
	}
	if proposal.Candidate.StableID != observation.Candidates[0].StableID || proposal.Candidate.InputDigest != observation.Candidates[0].InputDigest {
		return fmt.Errorf("proposal candidate is not the observed candidate")
	}
	if err := generation.ValidateSemanticAdoptionProposal(proposal); err != nil {
		return err
	}
	proposalDigest := cache.HashBytes(proposalData).String()
	if err := generation.ValidateSemanticAdoptionAuthorization(authorization); err != nil {
		return err
	}
	if authorization.ProposalDigest != proposalDigest || authorization.CandidateStableID != proposal.Candidate.StableID || authorization.CandidateInputDigest != proposal.Candidate.InputDigest ||
		authorization.ContractDigest != proposal.ContractDigest || authorization.InputSourceDigest != proposal.InputSourceDigest {
		return fmt.Errorf("authorization is not bound to the exact proposal candidate")
	}
	return nil
}

func runAdoptionPair(file *syntax.File, authorization generation.SemanticAdoptionAuthorization, adopted bool) (adoptionPair, error) {
	started := time.Now()
	var beforeMem, afterMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)
	pair := adoptionPair{operationCount: 2}
	var reuse semanticReuseCache
	for index := range 2 {
		var result generationResult
		var reused bool
		var err error
		if adopted {
			result, reused, err = generateWithDeadlineAdopted(file, nil, commandDeadline, authorization, &reuse)
		} else {
			result, err = generateWithDeadlineCore(file, nil, commandDeadline)
		}
		if err != nil {
			return adoptionPair{}, err
		}
		semanticDigest, err := cache.SemanticDigest(result.ir)
		if err != nil {
			return adoptionPair{}, fmt.Errorf("semantic adoption digest: %w", err)
		}
		if index == 0 {
			pair.first = result
			pair.firstSemanticDigest = semanticDigest
		} else {
			pair.replay = result
			pair.replaySemanticDigest = semanticDigest
		}
		if reused {
			pair.operationCount--
			pair.cacheHits++
		} else if adopted {
			pair.cacheMisses++
		}
	}
	runtime.ReadMemStats(&afterMem)
	pair.metrics = generation.SemanticAdoptionRuntimeMetrics{
		WallMS:          int64((time.Since(started).Nanoseconds() + int64(time.Millisecond) - 1) / int64(time.Millisecond)),
		AllocationCount: int64(afterMem.Mallocs - beforeMem.Mallocs),
		AllocationBytes: int64(afterMem.TotalAlloc - beforeMem.TotalAlloc),
	}
	return pair, nil
}

func adoptionEvidence(proposal generation.SemanticAdoptionProposal, proposalDigest, authorizationDigest string, before, adopted adoptionPair) generation.SemanticAdoptionEvidence {
	beforeSource := before.first.result.Source
	adoptedSource := adopted.first.result.Source
	beforeSemantic := before.firstSemanticDigest.String()
	adoptedSemantic := adopted.firstSemanticDigest.String()
	beforeDeterministic := bytes.Equal(beforeSource, before.replay.result.Source) && before.firstSemanticDigest == before.replaySemanticDigest
	adoptedDeterministic := bytes.Equal(adoptedSource, adopted.replay.result.Source) && adopted.firstSemanticDigest == adopted.replaySemanticDigest
	return generation.SemanticAdoptionEvidence{
		Schema: generation.SemanticAdoptionEvidenceSchema, ProposalDigest: proposalDigest,
		AuthorizationDigest: authorizationDigest, CandidateStableID: proposal.Candidate.StableID,
		InputDigest: proposal.Candidate.InputDigest, BeforeOperationCount: before.operationCount,
		AfterOperationCount: adopted.operationCount, CacheMisses: adopted.cacheMisses,
		CacheHits: adopted.cacheHits, ReuseApplied: adopted.operationCount < before.operationCount,
		BehaviorEqual:      bytes.Equal(beforeSource, adoptedSource) && beforeSemantic == adoptedSemantic,
		DeterminismEqual:   beforeDeterministic && adoptedDeterministic,
		BeforeOutputDigest: cache.HashBytes(beforeSource).String(), AdoptedOutputDigest: cache.HashBytes(adoptedSource).String(),
		AdoptedReplayDigest: cache.HashBytes(adopted.replay.result.Source).String(), BeforeSemanticDigest: beforeSemantic,
		AdoptedSemanticDigest: adoptedSemantic, Decision: "CLOSED", Reason: generation.SemanticAdoptionClosedReason,
		RepositoryWrites: 0, LocalTestExecutions: 0,
	}
}

func adoptionLifecycle(authorized bool) string {
	if authorized {
		return "AUTHORIZED_ADOPTION"
	}
	return "AUTHORIZATION_REQUIRED"
}
