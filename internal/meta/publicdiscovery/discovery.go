package publicdiscovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/discoverypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	Operation               = "gooo.generate.public-self-observation"
	ObservationLedgerSchema = "gooo/public-self-observation-ledger/v1"
	CandidateSchema         = "gooo/public-self-observation-candidate/v1"
	InvocationReceiptSchema = "gooo/public-self-observation-invocation/v1"
	ReportSchema            = "gooo/public-self-observation-report/v1"

	LedgerFilename     = "observation-ledger.ndjson"
	CandidateFilename  = "observation-candidate-%s.json"
	InvocationFilename = "invocation-receipt.json"
	MachineFilename    = "observation-report.json"
	HumanFilename      = "observation-report.md"

	ReasonSingle        = "INSUFFICIENT_COMPARABLE_OBSERVATIONS"
	ReasonIncompatible  = "CURRENT_GROUP_BELOW_QUORUM"
	ReasonTampered      = "TAMPERED_LEDGER_ENTRY"
	ReasonContradictory = "CONTRADICTORY_SEMANTIC_OR_OUTPUT_DIGEST"
	ReasonClosed        = "EXACT_COMPATIBLE_OBSERVATION_PAIR"

	Quorum = 2
)

// Input is the digest-only observation of one real gooo generate operation.
// Paths, source text, generated source, and environment values are deliberately
// excluded so that the ledger can be moved and compared across invocations.
type Input struct {
	SourceDigest            string
	InputSemanticDigest     string
	PreviousGoDigest        string
	ToolchainDigest         string
	ContractDigest          string
	EvaluatorDigest         string
	GeneratedSemanticDigest string
	GeneratedOutputDigest   string
	GeneratedManifestDigest string
}

type LedgerEntry struct {
	Schema                  string `json:"schema"`
	Sequence                int    `json:"sequence"`
	Operation               string `json:"operation"`
	SourceDigest            string `json:"source_digest"`
	InputSemanticDigest     string `json:"input_semantic_digest"`
	PreviousGoDigest        string `json:"previous_go_digest"`
	ToolchainDigest         string `json:"toolchain_digest"`
	ContractDigest          string `json:"contract_digest"`
	EvaluatorDigest         string `json:"evaluator_digest"`
	GeneratedSemanticDigest string `json:"generated_semantic_digest"`
	GeneratedOutputDigest   string `json:"generated_output_digest"`
	GeneratedManifestDigest string `json:"generated_manifest_digest"`
	RepositoryWrites        int    `json:"repository_writes"`
	LocalBuildExecutions    int    `json:"local_build_executions"`
	LocalTestExecutions     int    `json:"local_test_executions"`
	ObservationDigest       string `json:"observation_digest"`
}

type Candidate struct {
	Schema                  string `json:"schema"`
	CandidateID             string `json:"candidate_id"`
	Operation               string `json:"operation"`
	Decision                string `json:"decision"`
	Reason                  string `json:"reason"`
	GroupKeyDigest          string `json:"group_key_digest"`
	SourceDigest            string `json:"source_digest"`
	InputSemanticDigest     string `json:"input_semantic_digest"`
	PreviousGoDigest        string `json:"previous_go_digest"`
	ToolchainDigest         string `json:"toolchain_digest"`
	ContractDigest          string `json:"contract_digest"`
	EvaluatorDigest         string `json:"evaluator_digest"`
	GeneratedSemanticDigest string `json:"generated_semantic_digest"`
	GeneratedOutputDigest   string `json:"generated_output_digest"`
	GeneratedManifestDigest string `json:"generated_manifest_digest"`
	Quorum                  int    `json:"quorum"`
	ExecutionAllowed        bool   `json:"execution_allowed"`
	AuthorizationRequired   bool   `json:"authorization_required"`
	ProposalRequired        bool   `json:"proposal_required"`
	CertificateRequired     bool   `json:"certificate_required"`
	RepositoryWrites        int    `json:"repository_writes"`
	LocalBuildExecutions    int    `json:"local_build_executions"`
	LocalTestExecutions     int    `json:"local_test_executions"`
}

type InvocationReceipt struct {
	Schema                  string `json:"schema"`
	Sequence                int    `json:"sequence"`
	ObservationDigest       string `json:"observation_digest"`
	GroupKeyDigest          string `json:"group_key_digest"`
	Decision                string `json:"decision"`
	Reason                  string `json:"reason"`
	CandidateEmitted        bool   `json:"candidate_emitted"`
	LedgerLines             int    `json:"ledger_lines"`
	LedgerBytes             int    `json:"ledger_bytes"`
	ArtifactDenominator     int    `json:"artifact_denominator"`
	ArtifactCount           int    `json:"artifact_count"`
	RepositoryWrites        int    `json:"repository_writes"`
	LocalBuildExecutions    int    `json:"local_build_executions"`
	LocalTestExecutions     int    `json:"local_test_executions"`
	WallMS                  int64  `json:"wall_ms"`
	PeakRSSKib              int64  `json:"peak_rss_kib"`
	CandidateByteMismatches int    `json:"candidate_byte_mismatches"`
}

type Report struct {
	Schema                       string                           `json:"schema"`
	Decision                     string                           `json:"decision"`
	Reason                       string                           `json:"reason"`
	Unknown                      *generation.EnvelopeUnknownState `json:"unknown"`
	CaseIDs                      []string                         `json:"case_ids"`
	ObservationDigest            string                           `json:"observation_digest"`
	GroupKeyDigest               string                           `json:"group_key_digest"`
	SourceDigest                 string                           `json:"source_digest"`
	InputSemanticDigest          string                           `json:"input_semantic_digest"`
	PreviousGoDigest             string                           `json:"previous_go_digest"`
	ToolchainDigest              string                           `json:"toolchain_digest"`
	ContractDigest               string                           `json:"contract_digest"`
	EvaluatorDigest              string                           `json:"evaluator_digest"`
	GeneratedSemanticDigest      string                           `json:"generated_semantic_digest"`
	GeneratedOutputDigest        string                           `json:"generated_output_digest"`
	GeneratedManifestDigest      string                           `json:"generated_manifest_digest"`
	InvocationsObserved          int                              `json:"invocations_observed"`
	ComparableObservations       int                              `json:"comparable_observations"`
	IncompatibleObservations     int                              `json:"incompatible_observations"`
	ContradictoryObservations    int                              `json:"contradictory_observations"`
	DuplicateLedgerEntries       int                              `json:"duplicate_ledger_entries"`
	CandidatesEmitted            int                              `json:"candidates_emitted"`
	LedgerLines                  int                              `json:"ledger_lines"`
	LedgerBytes                  int                              `json:"ledger_bytes"`
	ArtifactDenominator          int                              `json:"artifact_denominator"`
	ArtifactCount                int                              `json:"artifact_count"`
	CandidatePath                string                           `json:"candidate_path,omitempty"`
	InvocationReceiptPath        string                           `json:"invocation_receipt_path"`
	MachineReportPath            string                           `json:"machine_report_path"`
	HumanReportPath              string                           `json:"human_report_path"`
	SemanticOperations           int                              `json:"semantic_operations"`
	CandidateByteMismatches      int                              `json:"candidate_byte_mismatches"`
	PolicyActivitiesExpected     int                              `json:"policy_activities_expected"`
	PolicyActivitiesObserved     int                              `json:"policy_activities_observed"`
	GeneratedEvaluatorMismatches int                              `json:"generated_evaluator_mismatches"`
	RepositoryWrites             int                              `json:"repository_writes"`
	LocalBuildExecutions         int                              `json:"local_build_executions"`
	LocalTestExecutions          int                              `json:"local_test_executions"`
	WallMS                       int64                            `json:"wall_ms"`
	PeakRSSKib                   int64                            `json:"peak_rss_kib"`
	CandidateBytesEqual          bool                             `json:"candidate_bytes_equal"`
	CandidateByteReplayEqual     bool                             `json:"candidate_byte_replay_equal"`
}

type Result struct {
	Report              Report
	Receipt             InvocationReceipt
	CandidatePath       string
	CandidateBytes      []byte
	CandidateByteReplay []byte
}

// GroupKey is the exact comparable identity shared with independent verifiers.
// Generated semantic/output digests are intentionally not part of this key;
// they are compared after compatible invocations have been grouped.
type GroupKey struct {
	Operation           string
	SourceDigest        string
	InputSemanticDigest string
	PreviousGoDigest    string
	ToolchainDigest     string
	ContractDigest      string
	EvaluatorDigest     string
}

type outputSignature struct {
	SemanticDigest string
	OutputDigest   string
	ManifestDigest string
}

func PolicySourceDigest() string       { return discoverypolicy.PolicySourceDigest() }
func GeneratedEvaluatorDigest() string { return discoverypolicy.GeneratedEvaluatorDigest() }

func Record(outputDir string, input Input, wallMS, peakRSSKib int64) (Result, error) {
	if outputDir == "" {
		return Result{}, errors.New("public discovery output directory is empty")
	}
	if err := validateInput(input); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create public discovery output: %w", err)
	}
	ledgerPath := filepath.Join(outputDir, LedgerFilename)
	rawLedger, readErr := os.ReadFile(ledgerPath)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return Result{}, fmt.Errorf("read public discovery ledger: %w", readErr)
	}
	entries, tampered, parseErr := parseLedger(rawLedger)
	ledgerLines := countLedgerLines(rawLedger)
	if parseErr != nil {
		return writeInvalidResult(outputDir, input, ledgerLines, len(rawLedger), wallMS, peakRSSKib, tampered, parseErr)
	}

	entry := LedgerEntry{
		Schema: ObservationLedgerSchema, Sequence: len(entries) + 1, Operation: Operation,
		SourceDigest: input.SourceDigest, InputSemanticDigest: input.InputSemanticDigest,
		PreviousGoDigest: input.PreviousGoDigest, ToolchainDigest: input.ToolchainDigest,
		ContractDigest: input.ContractDigest, EvaluatorDigest: input.EvaluatorDigest,
		GeneratedSemanticDigest: input.GeneratedSemanticDigest, GeneratedOutputDigest: input.GeneratedOutputDigest,
		GeneratedManifestDigest: input.GeneratedManifestDigest,
		RepositoryWrites:        0, LocalBuildExecutions: 0, LocalTestExecutions: 0,
	}
	entry.ObservationDigest, parseErr = entryContentDigest(entry)
	if parseErr != nil {
		return Result{}, parseErr
	}
	entryBytes, err := marshalJSONLine(entry)
	if err != nil {
		return Result{}, err
	}
	file, err := os.OpenFile(ledgerPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return Result{}, fmt.Errorf("open public discovery ledger: %w", err)
	}
	_, writeErr := file.Write(entryBytes)
	closeErr := file.Close()
	if writeErr != nil {
		return Result{}, fmt.Errorf("append public discovery ledger: %w", writeErr)
	}
	if closeErr != nil {
		return Result{}, fmt.Errorf("close public discovery ledger: %w", closeErr)
	}
	entries = append(entries, entry)
	rawLedger = append(rawLedger, entryBytes...)
	return writeResult(outputDir, entries, entry, ledgerLines+1, len(rawLedger), wallMS, peakRSSKib)
}

func writeInvalidResult(outputDir string, input Input, ledgerLines, ledgerBytes int, wallMS, peakRSSKib int64, tampered bool, parseErr error) (Result, error) {
	reason := ReasonTampered
	caseID := discoverypolicy.CaseTamperedEntry
	if !tampered {
		reason = ReasonTampered
	}
	entry := LedgerEntry{Schema: ObservationLedgerSchema, Sequence: 0, Operation: Operation,
		SourceDigest: input.SourceDigest, InputSemanticDigest: input.InputSemanticDigest,
		PreviousGoDigest: input.PreviousGoDigest, ToolchainDigest: input.ToolchainDigest,
		ContractDigest: input.ContractDigest, EvaluatorDigest: input.EvaluatorDigest,
		GeneratedSemanticDigest: input.GeneratedSemanticDigest, GeneratedOutputDigest: input.GeneratedOutputDigest,
		GeneratedManifestDigest: input.GeneratedManifestDigest}
	result := Result{}
	report := baseReport(outputDir, input, entry, []string{caseID}, "REFUTED", reason, ledgerLines, ledgerBytes, wallMS, peakRSSKib)
	report.Unknown = nil
	report.CandidateBytesEqual = false
	report.CandidateByteReplayEqual = false
	report.ContradictoryObservations = 0
	result.Report = report
	result.Receipt = receiptFromReport(report, 0, false)
	if err := writeReportArtifacts(outputDir, result, fmt.Sprintf("ledger validation failed: %v", parseErr)); err != nil {
		return Result{}, err
	}
	return result, nil
}

func writeResult(outputDir string, entries []LedgerEntry, current LedgerEntry, ledgerLines, ledgerBytes int, wallMS, peakRSSKib int64) (Result, error) {
	keyDigest, err := digestGroupKey(current)
	if err != nil {
		return Result{}, err
	}
	currentGroup := make([]LedgerEntry, 0, len(entries))
	incompatible := 0
	for _, entry := range entries {
		entryKey, digestErr := digestGroupKey(entry)
		if digestErr != nil {
			return Result{}, digestErr
		}
		if entryKey == keyDigest {
			currentGroup = append(currentGroup, entry)
		} else {
			incompatible++
		}
	}
	contradictory := countContradictory(currentGroup)
	caseIDs := make([]string, 0, 2)
	decision := discoverypolicy.DecisionUnknown
	reason := ReasonSingle
	unknown := &generation.EnvelopeUnknownState{
		Stage: "PUBLIC_SELF_OBSERVATION", Step: "COMPARE_CALLER_OWNED_LEDGER",
		Reason: ReasonSingle, UnknownClass: "INCOMPLETE_EVIDENCE",
		NextOperation: "RECORD_ONE_MORE_EXACT_COMPARABLE_GENERATE",
		BlockedBy:     []string{"comparable_observation_quorum"},
	}
	evaluatorMismatches := 0
	switch {
	case contradictory > 0:
		decision = discoverypolicy.DecisionRefuted
		reason = ReasonContradictory
		caseIDs = append(caseIDs, discoverypolicy.CaseContradictory)
		unknown = nil
	case len(currentGroup) < Quorum:
		if incompatible > 0 {
			reason = ReasonIncompatible
			unknown.Reason = ReasonIncompatible
			unknown.BlockedBy = []string{"same_source_input_toolchain_contract_group"}
			caseIDs = append(caseIDs, discoverypolicy.CaseIncompatibleGroup)
		} else {
			caseIDs = append(caseIDs, discoverypolicy.CaseSingleComparable)
		}
	case len(currentGroup) >= Quorum:
		// Additional exact observations remain safe evidence; the candidate
		// remains one stable result for this group.
		caseIDs = append(caseIDs, discoverypolicy.CaseExactPairCandidate, discoverypolicy.CaseDeterministicReplay)
		for _, caseID := range caseIDs {
			if expected, ok := discoverypolicy.Evaluate(caseID); !ok || expected != discoverypolicy.DecisionClosed {
				evaluatorMismatches++
			}
		}
		if evaluatorMismatches == 0 {
			decision = discoverypolicy.DecisionClosed
			reason = ReasonClosed
			unknown = nil
		} else {
			decision = discoverypolicy.DecisionRefuted
			reason = ReasonContradictory
			caseIDs = []string{discoverypolicy.CaseContradictory}
			unknown = nil
		}
	}
	if decision == discoverypolicy.DecisionUnknown {
		if expected, ok := discoverypolicy.Evaluate(caseIDs[0]); !ok || expected != decision {
			evaluatorMismatches++
		}
	}
	if decision == discoverypolicy.DecisionRefuted {
		if expected, ok := discoverypolicy.Evaluate(caseIDs[0]); !ok || expected != decision {
			evaluatorMismatches++
		}
	}
	candidatePath := ""
	candidateBytes := []byte(nil)
	replayBytes := []byte(nil)
	candidateMismatches := 0
	candidateEmitted := false
	if decision == discoverypolicy.DecisionClosed {
		candidate := Candidate{Schema: CandidateSchema,
			CandidateID: "public-generation-discovery/candidate/" + keyDigest,
			Operation:   Operation, Decision: discoverypolicy.DecisionClosed, Reason: reason,
			GroupKeyDigest: keyDigest, SourceDigest: current.SourceDigest,
			InputSemanticDigest: current.InputSemanticDigest, PreviousGoDigest: current.PreviousGoDigest,
			ToolchainDigest: current.ToolchainDigest, ContractDigest: current.ContractDigest,
			EvaluatorDigest: current.EvaluatorDigest, GeneratedSemanticDigest: current.GeneratedSemanticDigest,
			GeneratedOutputDigest: current.GeneratedOutputDigest, GeneratedManifestDigest: current.GeneratedManifestDigest,
			Quorum: Quorum, ExecutionAllowed: false, AuthorizationRequired: true, ProposalRequired: true, CertificateRequired: true,
			RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0}
		candidateBytes, err = marshalJSON(candidate)
		if err != nil {
			return Result{}, err
		}
		replayBytes, err = marshalJSON(candidate)
		if err != nil {
			return Result{}, err
		}
		candidatePath = filepath.Join(outputDir, fmt.Sprintf(CandidateFilename, keyDigest))
		if existing, readErr := os.ReadFile(candidatePath); readErr == nil {
			if !bytes.Equal(existing, candidateBytes) {
				candidateMismatches++
			}
		} else if !errors.Is(readErr, os.ErrNotExist) {
			return Result{}, fmt.Errorf("read public discovery candidate: %w", readErr)
		}
		if candidateMismatches == 0 {
			if err := os.WriteFile(candidatePath, candidateBytes, 0o644); err != nil {
				return Result{}, fmt.Errorf("write public discovery candidate: %w", err)
			}
			candidateEmitted = true
		}
		if !bytes.Equal(candidateBytes, replayBytes) {
			candidateMismatches++
		}
		if candidateMismatches > 0 {
			decision = discoverypolicy.DecisionRefuted
			reason = ReasonContradictory
			caseIDs = []string{discoverypolicy.CaseContradictory}
			candidateEmitted = false
		}
	}
	artifactCount := 4
	if candidateEmitted {
		artifactCount++
	}
	report := baseReport(outputDir, inputFromEntry(current), current, caseIDs, decision, reason, ledgerLines, ledgerBytes, wallMS, peakRSSKib)
	report.Unknown = unknown
	report.InvocationsObserved = len(entries)
	report.ComparableObservations = len(currentGroup)
	report.IncompatibleObservations = incompatible
	report.ContradictoryObservations = contradictory
	report.DuplicateLedgerEntries = maxInt(0, len(currentGroup)-1)
	report.CandidatesEmitted = boolInt(candidateEmitted)
	report.ArtifactDenominator = artifactCount
	report.ArtifactCount = artifactCount
	report.CandidatePath = candidatePath
	report.CandidateByteMismatches = candidateMismatches
	report.PolicyActivitiesObserved = 1
	report.GeneratedEvaluatorMismatches = evaluatorMismatches
	report.CandidateBytesEqual = decision == discoverypolicy.DecisionClosed && candidateMismatches == 0
	report.CandidateByteReplayEqual = bytes.Equal(candidateBytes, replayBytes) && len(candidateBytes) > 0
	result := Result{Report: report, CandidatePath: candidatePath, CandidateBytes: candidateBytes, CandidateByteReplay: replayBytes,
		Receipt: receiptFromReport(report, current.Sequence, candidateEmitted)}
	if err := writeReportArtifacts(outputDir, result, ""); err != nil {
		return Result{}, err
	}
	return result, nil
}

func baseReport(outputDir string, input Input, current LedgerEntry, caseIDs []string, decision, reason string, ledgerLines, ledgerBytes int, wallMS, peakRSSKib int64) Report {
	return Report{Schema: ReportSchema, Decision: decision, Reason: reason, CaseIDs: append([]string(nil), caseIDs...),
		ObservationDigest: current.ObservationDigest, GroupKeyDigest: digestGroupKeyMust(current),
		SourceDigest: input.SourceDigest, InputSemanticDigest: input.InputSemanticDigest, PreviousGoDigest: input.PreviousGoDigest,
		ToolchainDigest: input.ToolchainDigest, ContractDigest: input.ContractDigest, EvaluatorDigest: input.EvaluatorDigest,
		GeneratedSemanticDigest: input.GeneratedSemanticDigest, GeneratedOutputDigest: input.GeneratedOutputDigest,
		GeneratedManifestDigest: input.GeneratedManifestDigest, InvocationsObserved: 1, ComparableObservations: 1,
		LedgerLines: ledgerLines, LedgerBytes: ledgerBytes, ArtifactDenominator: 4, ArtifactCount: 4,
		InvocationReceiptPath: filepath.Join(outputDir, InvocationFilename), MachineReportPath: filepath.Join(outputDir, MachineFilename),
		HumanReportPath: filepath.Join(outputDir, HumanFilename), SemanticOperations: 1,
		PolicyActivitiesExpected: 1, PolicyActivitiesObserved: 1, RepositoryWrites: 0, LocalBuildExecutions: 0,
		LocalTestExecutions: 0, WallMS: wallMS, PeakRSSKib: peakRSSKib}
}

func inputFromEntry(entry LedgerEntry) Input {
	return Input{SourceDigest: entry.SourceDigest, InputSemanticDigest: entry.InputSemanticDigest,
		PreviousGoDigest: entry.PreviousGoDigest, ToolchainDigest: entry.ToolchainDigest,
		ContractDigest: entry.ContractDigest, EvaluatorDigest: entry.EvaluatorDigest,
		GeneratedSemanticDigest: entry.GeneratedSemanticDigest, GeneratedOutputDigest: entry.GeneratedOutputDigest,
		GeneratedManifestDigest: entry.GeneratedManifestDigest}
}

func receiptFromReport(report Report, sequence int, candidate bool) InvocationReceipt {
	return InvocationReceipt{Schema: InvocationReceiptSchema, Sequence: sequence, ObservationDigest: report.ObservationDigest,
		GroupKeyDigest: report.GroupKeyDigest, Decision: report.Decision, Reason: report.Reason,
		CandidateEmitted: candidate, LedgerLines: report.LedgerLines, LedgerBytes: report.LedgerBytes,
		ArtifactDenominator: report.ArtifactDenominator, ArtifactCount: report.ArtifactCount,
		RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0,
		WallMS: report.WallMS, PeakRSSKib: report.PeakRSSKib, CandidateByteMismatches: report.CandidateByteMismatches}
}

func writeReportArtifacts(outputDir string, result Result, note string) error {
	reportBytes, err := marshalJSON(result.Report)
	if err != nil {
		return err
	}
	receiptBytes, err := marshalJSON(result.Receipt)
	if err != nil {
		return err
	}
	human := renderHumanReport(result.Report, note)
	if err := os.WriteFile(filepath.Join(outputDir, MachineFilename), reportBytes, 0o644); err != nil {
		return fmt.Errorf("write public discovery report: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, InvocationFilename), receiptBytes, 0o644); err != nil {
		return fmt.Errorf("write public discovery receipt: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, HumanFilename), []byte(human), 0o644); err != nil {
		return fmt.Errorf("write public discovery human report: %w", err)
	}
	return nil
}

func renderHumanReport(report Report, note string) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Gooo public self-observation\n\nDecision: `%s`\nReason: `%s`\n\n", report.Decision, report.Reason)
	fmt.Fprintf(&builder, "Comparable observations: `%d`\nInvocations observed: `%d`\nCandidates emitted: `%d`\nArtifacts: `%d`\n", report.ComparableObservations, report.InvocationsObserved, report.CandidatesEmitted, report.ArtifactCount)
	if report.Unknown != nil {
		fmt.Fprintf(&builder, "\nEvidence remains bounded: record one more exact comparable generate invocation.\nBlocked by: `%s`\n", strings.Join(report.Unknown.BlockedBy, "`, `"))
	}
	if report.CandidatePath != "" {
		fmt.Fprintf(&builder, "\nCandidate: `%s`\nExecution allowed: `false`\nAuthorization required: `true`\n", report.CandidatePath)
	}
	if note != "" {
		fmt.Fprintf(&builder, "\nNote: %s\n", note)
	}
	fmt.Fprintf(&builder, "\nRepository writes: `%d`\nLocal build executions: `%d`\nLocal test executions: `%d`\n", report.RepositoryWrites, report.LocalBuildExecutions, report.LocalTestExecutions)
	return builder.String()
}

func parseLedger(raw []byte) ([]LedgerEntry, bool, error) {
	if len(raw) == 0 {
		return nil, false, nil
	}
	if raw[len(raw)-1] != '\n' {
		return nil, true, errors.New("ledger does not end with a newline")
	}
	lines := bytes.Split(raw, []byte{'\n'})
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	entries := make([]LedgerEntry, 0, len(lines))
	for index, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			return entries, true, fmt.Errorf("ledger line %d is blank", index+1)
		}
		var entry LedgerEntry
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&entry); err != nil {
			return entries, true, fmt.Errorf("ledger line %d: %w", index+1, err)
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			if err == nil {
				return entries, true, fmt.Errorf("ledger line %d contains multiple JSON values", index+1)
			}
			return entries, true, fmt.Errorf("ledger line %d: %w", index+1, err)
		}
		if err := validateEntry(entry, index+1); err != nil {
			return entries, true, err
		}
		digest, err := entryContentDigest(entry)
		if err != nil {
			return entries, true, err
		}
		if digest != entry.ObservationDigest {
			return entries, true, fmt.Errorf("ledger line %d observation digest does not match content", index+1)
		}
		entries = append(entries, entry)
	}
	return entries, false, nil
}

func validateInput(input Input) error {
	values := []string{input.SourceDigest, input.InputSemanticDigest, input.PreviousGoDigest, input.ToolchainDigest,
		input.ContractDigest, input.EvaluatorDigest, input.GeneratedSemanticDigest, input.GeneratedOutputDigest, input.GeneratedManifestDigest}
	for index, value := range values {
		if !cache.Digest(value).Known() {
			return fmt.Errorf("public discovery input digest %d is not a known SHA-256 digest", index+1)
		}
	}
	return nil
}

func validateEntry(entry LedgerEntry, wantSequence int) error {
	if entry.Schema != ObservationLedgerSchema || entry.Operation != Operation {
		return fmt.Errorf("ledger line %d schema or operation is invalid", wantSequence)
	}
	if entry.Sequence != wantSequence {
		return fmt.Errorf("ledger line %d sequence = %d, want %d", wantSequence, entry.Sequence, wantSequence)
	}
	if entry.RepositoryWrites != 0 || entry.LocalBuildExecutions != 0 || entry.LocalTestExecutions != 0 {
		return fmt.Errorf("ledger line %d claims a forbidden effect", wantSequence)
	}
	return validateInput(Input{SourceDigest: entry.SourceDigest, InputSemanticDigest: entry.InputSemanticDigest,
		PreviousGoDigest: entry.PreviousGoDigest, ToolchainDigest: entry.ToolchainDigest, ContractDigest: entry.ContractDigest,
		EvaluatorDigest: entry.EvaluatorDigest, GeneratedSemanticDigest: entry.GeneratedSemanticDigest,
		GeneratedOutputDigest: entry.GeneratedOutputDigest, GeneratedManifestDigest: entry.GeneratedManifestDigest})
}

func entryContentDigest(entry LedgerEntry) (string, error) {
	entry.ObservationDigest = ""
	digest, err := cache.DigestOf(entry)
	if err != nil {
		return "", fmt.Errorf("public discovery observation digest: %w", err)
	}
	return digest.String(), nil
}

func GroupKeyDigest(entry LedgerEntry) (string, error) {
	digest, err := cache.DigestOf(GroupKey{Operation: entry.Operation, SourceDigest: entry.SourceDigest,
		InputSemanticDigest: entry.InputSemanticDigest, PreviousGoDigest: entry.PreviousGoDigest,
		ToolchainDigest: entry.ToolchainDigest, ContractDigest: entry.ContractDigest, EvaluatorDigest: entry.EvaluatorDigest})
	if err != nil {
		return "", fmt.Errorf("public discovery group digest: %w", err)
	}
	return digest.String(), nil
}

func EntryContentDigest(entry LedgerEntry) (string, error) { return entryContentDigest(entry) }

func digestGroupKey(entry LedgerEntry) (string, error) { return GroupKeyDigest(entry) }

func digestGroupKeyMust(entry LedgerEntry) string {
	digest, _ := digestGroupKey(entry)
	return digest
}

func countContradictory(entries []LedgerEntry) int {
	if len(entries) < 2 {
		return 0
	}
	first := outputSignature{entries[0].GeneratedSemanticDigest, entries[0].GeneratedOutputDigest, entries[0].GeneratedManifestDigest}
	count := 0
	for _, entry := range entries[1:] {
		if (outputSignature{entry.GeneratedSemanticDigest, entry.GeneratedOutputDigest, entry.GeneratedManifestDigest}) != first {
			count++
		}
	}
	return count
}

func marshalJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func marshalJSONLine(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func countLedgerLines(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	count := bytes.Count(raw, []byte{'\n'})
	if raw[len(raw)-1] != '\n' {
		count++
	}
	return count
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
