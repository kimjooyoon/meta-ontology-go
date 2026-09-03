package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type retentionEvidenceOptions struct {
	observationFilename   string
	proposalFilename      string
	authorizationFilename string
	adoptionFilename      string
}

type retentionEvidence struct {
	observationData   []byte
	observation       generation.SemanticObservation
	proposalData      []byte
	proposal          generation.SemanticAdoptionProposal
	authorizationData []byte
	authorization     generation.SemanticAdoptionAuthorization
	adoptionData      []byte
	adoption          generation.SemanticAdoptionReport
}

type retentionInputBindingError struct{ message string }

func (err retentionInputBindingError) Error() string { return err.message }

func readRetentionEvidence(options retentionEvidenceOptions, reader SourceReader) (retentionEvidence, error) {
	var evidence retentionEvidence
	var err error
	evidence.observationData, err = reader.ReadFile(options.observationFilename)
	if err != nil {
		return evidence, fmt.Errorf("read observation: %w", err)
	}
	if err := json.Unmarshal(evidence.observationData, &evidence.observation); err != nil {
		return evidence, fmt.Errorf("decode observation: %w", err)
	}
	evidence.proposalData, err = reader.ReadFile(options.proposalFilename)
	if err != nil {
		return evidence, fmt.Errorf("read proposal: %w", err)
	}
	if err := json.Unmarshal(evidence.proposalData, &evidence.proposal); err != nil {
		return evidence, fmt.Errorf("decode proposal: %w", err)
	}
	evidence.adoptionData, err = reader.ReadFile(options.adoptionFilename)
	if err != nil {
		return evidence, fmt.Errorf("read adoption result: %w", err)
	}
	if err := json.Unmarshal(evidence.adoptionData, &evidence.adoption); err != nil {
		return evidence, fmt.Errorf("decode adoption result: %w", err)
	}
	if options.authorizationFilename != "" {
		evidence.authorizationData, err = reader.ReadFile(options.authorizationFilename)
		if err != nil {
			return evidence, fmt.Errorf("read authorization: %w", err)
		}
		if err := json.Unmarshal(evidence.authorizationData, &evidence.authorization); err != nil {
			return evidence, fmt.Errorf("decode authorization: %w", err)
		}
	}
	return evidence, nil
}

func validateRetentionEvidence(inputs observationInputs, evidence retentionEvidence) error {
	if err := generation.VerifySemanticObservation(evidence.observation); err != nil {
		return fmt.Errorf("observation verification: %w", err)
	}
	if evidence.observation.Decision != "CLOSED" || len(evidence.observation.Candidates) != 1 {
		return errors.New("observation must be CLOSED with one candidate")
	}
	if evidence.observation.ContractDigest != cache.HashBytes(inputs.contractSource).String() {
		return retentionInputBindingError{"observation is not bound to the supplied contract"}
	}
	if evidence.observation.InputSourceDigest != cache.HashBytes(inputs.inputSource).String() {
		return retentionInputBindingError{"observation is not bound to the supplied input"}
	}
	if err := generation.ValidateSemanticAdoptionProposal(evidence.proposal); err != nil {
		return err
	}
	if evidence.proposal.ObservationDigest != cache.HashBytes(evidence.observationData).String() {
		return errors.New("proposal does not bind the exact observation")
	}
	if evidence.proposal.ContractDigest != evidence.observation.ContractDigest || evidence.proposal.InputSourceDigest != evidence.observation.InputSourceDigest ||
		!reflect.DeepEqual(evidence.proposal.Candidate, evidence.observation.Candidates[0]) {
		return errors.New("proposal does not bind the observed candidate")
	}
	if evidence.proposal.ContractDigest != cache.HashBytes(inputs.contractSource).String() || evidence.proposal.InputSourceDigest != cache.HashBytes(inputs.inputSource).String() {
		return retentionInputBindingError{"proposal is not bound to the supplied contract and input"}
	}
	if err := validateRetentionAdoptionReport(evidence); err != nil {
		return err
	}
	return nil
}

func validateRetentionAdoptionReport(evidence retentionEvidence) error {
	proposalDigest := cache.HashBytes(evidence.proposalData).String()
	authorizationDigest := evidence.adoption.AuthorizationDigest
	observationDigest := cache.HashBytes(evidence.observationData).String()
	if evidence.adoption.Schema != generation.SemanticAdoptionReportSchema || evidence.adoption.Lifecycle != "AUTHORIZED_ADOPTION" ||
		evidence.adoption.ObservationDigest != observationDigest || evidence.adoption.ProposalDigest != proposalDigest ||
		evidence.adoption.AuthorizationDigest != authorizationDigest || !reflect.DeepEqual(evidence.adoption.Proposal, evidence.proposal) ||
		evidence.adoption.IndependentDecision != "CLOSED" ||
		evidence.adoption.IndependentReason != evidence.adoption.Evidence.Reason {
		return errors.New("adoption result is not a closed report bound to its evidence")
	}
	if err := generation.ValidateSemanticAdoptionAuthorization(evidence.adoption.Authorization); err != nil {
		return err
	}
	if decision, reason, _, err := generation.VerifySemanticAdoption(evidence.proposal, proposalDigest, evidence.adoption.Authorization, authorizationDigest, evidence.adoption.Evidence); err != nil || decision != "CLOSED" || reason != generation.SemanticAdoptionClosedReason {
		if err != nil {
			return fmt.Errorf("adoption verification: %w", err)
		}
		return fmt.Errorf("adoption verification = %s/%s", decision, reason)
	}
	if err := generation.ValidateBoundSemanticAdoption(evidence.adoption.Observation, evidence.proposal, evidence.adoption.Evidence); err != nil {
		return err
	}
	if err := generation.VerifySemanticObservation(evidence.adoption.Observation); err != nil {
		return fmt.Errorf("adopted observation verification: %w", err)
	}
	return nil
}

func validateRetentionAuthorization(evidence retentionEvidence) error {
	if len(evidence.authorizationData) == 0 {
		return errors.New("authorization is missing")
	}
	if err := generation.ValidateSemanticAdoptionAuthorization(evidence.authorization); err != nil {
		return err
	}
	if evidence.authorization.ProposalDigest != cache.HashBytes(evidence.proposalData).String() ||
		evidence.authorization.CandidateStableID != evidence.proposal.Candidate.StableID ||
		evidence.authorization.CandidateInputDigest != evidence.proposal.Candidate.InputDigest ||
		evidence.authorization.ContractDigest != evidence.proposal.ContractDigest ||
		evidence.authorization.InputSourceDigest != evidence.proposal.InputSourceDigest {
		return errors.New("authorization is not bound to the exact proposal candidate")
	}
	if evidence.authorization != evidence.adoption.Authorization {
		return errors.New("authorization differs from the authorized adoption report")
	}
	if cache.HashBytes(evidence.authorizationData).String() != evidence.adoption.AuthorizationDigest {
		return errors.New("authorization digest differs from the authorized adoption report")
	}
	return nil
}

func retentionBindings(inputs observationInputs, evidence retentionEvidence, compilerDigest, verifierDigest string) generation.SemanticRetentionBindings {
	contractDigest := cache.HashBytes(inputs.contractSource).String()
	return generation.SemanticRetentionBindings{
		AdoptionReportDigest: cache.HashBytes(evidence.adoptionData).String(),
		ObservationDigest:    cache.HashBytes(evidence.observationData).String(),
		ProposalDigest:       cache.HashBytes(evidence.proposalData).String(),
		AuthorizationDigest:  cache.HashBytes(evidence.authorizationData).String(),
		CandidateStableID:    evidence.proposal.Candidate.StableID,
		ContractSourceDigest: contractDigest,
		InputSourceDigest:    cache.HashBytes(inputs.inputSource).String(),
		NormalizedIRDigest:   evidence.proposal.Candidate.InputDigest,
		CompilerDigest:       compilerDigest,
		ToolchainDigest:      generation.SemanticRetentionToolchainDigest(),
		VerifierDigest:       verifierDigest,
		PolicyDigest:         contractDigest,
	}
}

func retentionResultBase(inputs observationInputs, evidence retentionEvidence, compilerDigest, verifierDigest string) generation.SemanticRetentionResult {
	result := generation.SemanticRetentionResult{
		Schema:               generation.SemanticRetentionResultSchema,
		Lifecycle:            "RETAINED_KNOWLEDGE",
		AdoptionReportDigest: cache.HashBytes(evidence.adoptionData).String(),
		ObservationDigest:    cache.HashBytes(evidence.observationData).String(),
		ProposalDigest:       cache.HashBytes(evidence.proposalData).String(),
		CandidateStableID:    evidence.proposal.Candidate.StableID,
		ContractSourceDigest: cache.HashBytes(inputs.contractSource).String(),
		InputSourceDigest:    cache.HashBytes(inputs.inputSource).String(),
		CompilerDigest:       compilerDigest,
		ToolchainDigest:      generation.SemanticRetentionToolchainDigest(),
		VerifierDigest:       verifierDigest,
		PolicyDigest:         cache.HashBytes(inputs.contractSource).String(),
		RepositoryWrites:     0,
		LocalTestExecutions:  0,
	}
	if len(evidence.authorizationData) != 0 {
		result.AuthorizationDigest = cache.HashBytes(evidence.authorizationData).String()
	}
	return result
}

func retentionUnknownResult(base generation.SemanticRetentionResult, reason, lifecycle string) generation.SemanticRetentionResult {
	base.Lifecycle = lifecycle
	base.Decision = "UNKNOWN"
	base.Reason = reason
	base.Unknown = generation.SemanticRetentionUnknownState(reason)
	base.Metrics = generation.SemanticRetentionRuntimeMetrics{}
	return base
}

func retentionRefutedResult(base generation.SemanticRetentionResult, certificateDigest string) generation.SemanticRetentionResult {
	base.Decision = generation.SemanticRetentionRefuted
	base.Reason = generation.SemanticRetentionRefutedReason
	base.Unknown = nil
	base.CertificateDigest = certificateDigest
	base.Metrics.CertificateMisses = 1
	return base
}

func retentionWallMS(started time.Time) int64 {
	return int64((time.Since(started).Nanoseconds() + int64(time.Millisecond) - 1) / int64(time.Millisecond))
}

func retentionRuntimeMetrics(started time.Time, before, after runtime.MemStats) generation.SemanticRetentionRuntimeMetrics {
	return generation.SemanticRetentionRuntimeMetrics{
		WallMS:          retentionWallMS(started),
		PeakRSSKib:      readPeakRSSKib(),
		AllocationCount: int64(after.Mallocs - before.Mallocs),
		AllocationBytes: int64(after.TotalAlloc - before.TotalAlloc),
	}
}

func readPeakRSSKib() int64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if !strings.HasPrefix(line, "VmHWM:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || value < 0 {
			return 0
		}
		return value
	}
	return 0
}

func writeRetentionResult(outputDir string, result generation.SemanticRetentionResult, stdout, stderr io.Writer) int {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention: encode result: %v\n", err)
		return exitFailure
	}
	data = append(data, '\n')
	if err := writeRetentionArtifacts(outputDir, []retentionArtifact{{name: "retention-result.json", data: data}}); err != nil {
		fmt.Fprintf(stderr, "gooo: retention: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "retention: %s (%s)\n", filepath.Join(outputDir, "retention-result.json"), result.Decision)
	return exitOK
}

type retentionArtifact struct {
	name string
	data []byte
}

func writeRetentionArtifacts(outputDir string, artifacts []retentionArtifact) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create caller-owned output directory: %w", err)
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("inspect caller-owned output directory: %w", err)
	}
	if len(entries) != 0 {
		return errors.New("caller-owned output directory must be empty")
	}
	for _, artifact := range artifacts {
		if artifact.name == "" || len(artifact.data) == 0 {
			return errors.New("retention artifact is empty")
		}
		if err := os.WriteFile(filepath.Join(outputDir, artifact.name), artifact.data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", artifact.name, err)
		}
	}
	return nil
}

func retentionResultEqualBytes(left, right generation.SemanticRetentionResult) bool {
	return bytes.Equal(left.GeneratedSource, right.GeneratedSource)
}
