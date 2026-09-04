package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/discoverypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type fixtureManifest struct {
	Schema string        `json:"schema"`
	Cases  []fixtureCase `json:"cases"`
}

type fixtureCase struct {
	ID        string `json:"id"`
	Report    string `json:"report"`
	Ledger    string `json:"ledger"`
	Candidate string `json:"candidate,omitempty"`
}

type verificationCase struct {
	ID                        string `json:"id"`
	ExpectedDecision          string `json:"expected_decision"`
	ObservedDecision          string `json:"observed_decision"`
	ExpectedReason            string `json:"expected_reason"`
	ObservedReason            string `json:"observed_reason"`
	LedgerValid               bool   `json:"ledger_valid"`
	CandidateBytesEqual       bool   `json:"candidate_bytes_equal"`
	CandidateReplayEqual      bool   `json:"candidate_replay_equal"`
	ComparableObservations    int    `json:"comparable_observations"`
	IncompatibleObservations  int    `json:"incompatible_observations"`
	ContradictoryObservations int    `json:"contradictory_observations"`
	CandidateByteMismatches   int    `json:"candidate_byte_mismatches"`
	Error                     string `json:"error,omitempty"`
}

type verificationReport struct {
	Schema                       string             `json:"schema"`
	ContractSourceDigest         string             `json:"contract_source_digest"`
	EvaluatorDigest              string             `json:"evaluator_digest"`
	CaseDenominator              int                `json:"case_denominator"`
	CaseIDs                      []string           `json:"case_ids"`
	Cases                        []verificationCase `json:"cases"`
	ClosedCases                  int                `json:"closed_cases"`
	UnknownCases                 int                `json:"unknown_cases"`
	RefutedCases                 int                `json:"refuted_cases"`
	PolicyActivitiesExpected     int                `json:"policy_activities_expected"`
	PolicyActivitiesObserved     int                `json:"policy_activities_observed"`
	GeneratedEvaluatorMismatches int                `json:"generated_evaluator_mismatches"`
	MalformedLedgerEntries       int                `json:"malformed_ledger_entries"`
	TamperedLedgerEntries        int                `json:"tampered_ledger_entries"`
	ContradictoryOutputs         int                `json:"contradictory_outputs"`
	CandidateByteMismatches      int                `json:"candidate_byte_mismatches"`
	RepositoryWrites             int                `json:"repository_writes"`
	LocalBuildExecutions         int                `json:"local_build_executions"`
	LocalTestExecutions          int                `json:"local_test_executions"`
}

type rawDecisionCase struct {
	ID       string
	Decision string
}

var fixedCaseIDs = []string{
	discoverypolicy.CaseExactPairCandidate,
	discoverypolicy.CaseDeterministicReplay,
	discoverypolicy.CaseSingleComparable,
	discoverypolicy.CaseIncompatibleGroup,
	discoverypolicy.CaseTamperedEntry,
	discoverypolicy.CaseContradictory,
}

func main() {
	contract := flag.String("contract", "", "authoritative discovery policy .gooo")
	manifest := flag.String("manifest", "", "verification fixture manifest")
	output := flag.String("output", "", "verification report output")
	forgeInput := flag.String("forge-contradiction", "", "valid ledger to copy and contradict")
	forgeOutput := flag.String("forge-output", "", "contradictory ledger output")
	flag.Parse()
	if *forgeInput != "" {
		if *forgeOutput == "" {
			exitError("-forge-output is required with -forge-contradiction")
		}
		if err := forgeContradiction(*forgeInput, *forgeOutput); err != nil {
			exitError(err.Error())
		}
		return
	}
	if *contract == "" || *manifest == "" || *output == "" {
		exitError("usage: self-improvement-public-discovery -contract FILE -manifest FILE -output FILE")
	}
	if err := verify(*contract, *manifest, *output); err != nil {
		exitError(err.Error())
	}
}

func verify(contractPath, manifestPath, outputPath string) error {
	contractSource, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("read contract: %w", err)
	}
	rows, activities, err := rawPolicyRows(contractPath, contractSource)
	if err != nil {
		return err
	}
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read fixture manifest: %w", err)
	}
	var manifest fixtureManifest
	if err := decodeStrict(manifestData, &manifest); err != nil {
		return fmt.Errorf("decode fixture manifest: %w", err)
	}
	if manifest.Schema != "gooo/public-self-observation-verification-input/v1" {
		return fmt.Errorf("fixture manifest schema %q is invalid", manifest.Schema)
	}
	if len(rows) != 6 || len(manifest.Cases) != 6 {
		return fmt.Errorf("verification boundary requires six policy rows and six fixtures")
	}
	rowByID := make(map[string]rawDecisionCase, len(rows))
	caseIDs := make([]string, 0, len(rows))
	generatedMismatches := 0
	for _, row := range rows {
		rowByID[row.ID] = row
		caseIDs = append(caseIDs, row.ID)
		decision, ok := discoverypolicy.Evaluate(row.ID)
		if !ok || decision != row.Decision {
			generatedMismatches++
		}
	}
	if !sameStrings(caseIDs, fixedCaseIDs) {
		return fmt.Errorf("raw policy case IDs do not preserve the fixed six-case boundary")
	}
	if activities != 1 {
		return fmt.Errorf("policy activity count = %d, want 1", activities)
	}
	report := verificationReport{Schema: "gooo/public-self-observation-verification/v1",
		ContractSourceDigest: cache.HashBytes(contractSource).String(), EvaluatorDigest: discoverypolicy.GeneratedEvaluatorDigest(),
		CaseDenominator: 6, CaseIDs: caseIDs, Cases: make([]verificationCase, 0, len(manifest.Cases)),
		PolicyActivitiesExpected: 1, PolicyActivitiesObserved: activities, GeneratedEvaluatorMismatches: generatedMismatches}
	for _, fixture := range manifest.Cases {
		row, ok := rowByID[fixture.ID]
		if !ok {
			return fmt.Errorf("fixture %q is outside the raw policy", fixture.ID)
		}
		result, err := verifyFixture(fixture, row.Decision, report.ContractSourceDigest, report.EvaluatorDigest)
		if err != nil {
			return err
		}
		report.Cases = append(report.Cases, result)
		switch result.ObservedDecision {
		case discoverypolicy.DecisionClosed:
			report.ClosedCases++
		case discoverypolicy.DecisionUnknown:
			report.UnknownCases++
		case discoverypolicy.DecisionRefuted:
			report.RefutedCases++
		}
		if !result.LedgerValid {
			report.MalformedLedgerEntries++
			report.TamperedLedgerEntries++
		}
		report.ContradictoryOutputs += result.ContradictoryObservations
		report.CandidateByteMismatches += result.CandidateByteMismatches
	}
	seenFixtures := make(map[string]bool, len(manifest.Cases))
	for _, fixture := range manifest.Cases {
		if seenFixtures[fixture.ID] {
			return fmt.Errorf("fixture %q is duplicated", fixture.ID)
		}
		seenFixtures[fixture.ID] = true
	}
	for _, caseID := range fixedCaseIDs {
		if !seenFixtures[caseID] {
			return fmt.Errorf("fixture %q is missing", caseID)
		}
	}
	if len(report.Cases) != 6 || report.ClosedCases != 2 || report.UnknownCases != 2 || report.RefutedCases != 2 {
		return fmt.Errorf("verification decisions do not preserve the 2/2/2 boundary")
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write verification report: %w", err)
	}
	return nil
}

func rawPolicyRows(filename string, source []byte) ([]rawDecisionCase, int, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return nil, 0, fmt.Errorf("parse policy: %w", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return nil, 0, fmt.Errorf("lower policy: %w", err)
	}
	rows := make([]rawDecisionCase, 0, 6)
	activities := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != discoverypolicy.PolicyActivity {
			continue
		}
		activities++
		seenHead := false
		for part := range strings.SplitSeq(node.ValueProgram, ";") {
			if part == discoverypolicy.DiscoveryDecisionHead {
				if seenHead {
					return nil, 0, errors.New("policy decision header is duplicated")
				}
				seenHead = true
				continue
			}
			if !strings.HasPrefix(part, "discovery-case=") {
				continue
			}
			encoded := strings.TrimPrefix(part, "discovery-case=")
			id, decision, ok := strings.Cut(encoded, ":")
			if !ok || id == "" || decision == "" || strings.Contains(decision, ":") {
				return nil, 0, fmt.Errorf("malformed raw policy row %q", encoded)
			}
			rows = append(rows, rawDecisionCase{ID: id, Decision: decision})
		}
		if !seenHead {
			return nil, 0, errors.New("raw policy decision header is missing")
		}
	}
	return rows, activities, nil
}

func verifyFixture(fixture fixtureCase, expectedDecision, contractDigest, evaluatorDigest string) (verificationCase, error) {
	reportData, err := os.ReadFile(fixture.Report)
	if err != nil {
		return verificationCase{}, fmt.Errorf("read report %q: %w", fixture.ID, err)
	}
	var observed publicdiscovery.Report
	if err := decodeStrict(reportData, &observed); err != nil {
		return verificationCase{}, fmt.Errorf("decode report %q: %w", fixture.ID, err)
	}
	result := verificationCase{ID: fixture.ID, ExpectedDecision: expectedDecision, ObservedDecision: observed.Decision,
		ExpectedReason: expectedReason(fixture.ID), ObservedReason: observed.Reason, CandidateByteMismatches: observed.CandidateByteMismatches}
	rawLedger, err := os.ReadFile(fixture.Ledger)
	if err != nil {
		return result, fmt.Errorf("read ledger %q: %w", fixture.ID, err)
	}
	receiptData, err := os.ReadFile(observed.InvocationReceiptPath)
	if err != nil {
		return result, fmt.Errorf("read receipt %q: %w", fixture.ID, err)
	}
	var receipt publicdiscovery.InvocationReceipt
	if err := decodeStrict(receiptData, &receipt); err != nil {
		return result, fmt.Errorf("decode receipt %q: %w", fixture.ID, err)
	}
	entries, ledgerErr := independentlyParseLedger(rawLedger)
	if ledgerErr != nil {
		result.LedgerValid = false
		if expectedDecision != discoverypolicy.DecisionRefuted || observed.Decision != discoverypolicy.DecisionRefuted || observed.Reason != publicdiscovery.ReasonTampered || !sameStrings(observed.CaseIDs, []string{discoverypolicy.CaseTamperedEntry}) {
			result.Error = ledgerErr.Error()
			return result, fmt.Errorf("fixture %q tampered ledger classification: %s", fixture.ID, result.Error)
		}
		if observed.CandidateByteMismatches != 0 || observed.ArtifactDenominator != 4 || observed.ArtifactCount != 4 {
			return result, fmt.Errorf("fixture %q tampered report is not fail-closed", fixture.ID)
		}
		if receipt.Schema != publicdiscovery.InvocationReceiptSchema || receipt.Sequence != 0 || receipt.Decision != observed.Decision || receipt.Reason != observed.Reason || receipt.LedgerLines != observed.LedgerLines || receipt.LedgerBytes != observed.LedgerBytes || receipt.CandidateEmitted || receipt.ArtifactDenominator != 4 || receipt.ArtifactCount != 4 {
			return result, fmt.Errorf("fixture %q tampered receipt is not fail-closed", fixture.ID)
		}
		return result, nil
	}
	result.LedgerValid = true
	if len(entries) == 0 {
		return result, errors.New("valid fixture ledger is empty")
	}
	current := entries[len(entries)-1]
	keyDigest, err := publicdiscovery.GroupKeyDigest(current)
	if err != nil {
		return result, err
	}
	group := make([]publicdiscovery.LedgerEntry, 0, len(entries))
	incompatible := 0
	for _, entry := range entries {
		entryKey, digestErr := publicdiscovery.GroupKeyDigest(entry)
		if digestErr != nil {
			return result, digestErr
		}
		if entryKey == keyDigest {
			group = append(group, entry)
		} else {
			incompatible++
		}
	}
	contradictory := contradictoryCount(group)
	decision, reason, caseIDs := expectedClassification(len(group), incompatible, contradictory)
	if decision != expectedDecision {
		return result, fmt.Errorf("fixture %q raw ledger outcome does not match raw policy outcome", fixture.ID)
	}
	result.ComparableObservations = len(group)
	result.IncompatibleObservations = incompatible
	result.ContradictoryObservations = contradictory
	if observed.Decision != decision || observed.Reason != reason || !sameStrings(observed.CaseIDs, caseIDs) {
		return result, fmt.Errorf("fixture %q report classification differs from raw ledger", fixture.ID)
	}
	if observed.ObservationDigest != current.ObservationDigest || observed.GroupKeyDigest != keyDigest {
		return result, fmt.Errorf("fixture %q report identity differs from raw ledger", fixture.ID)
	}
	if decision == discoverypolicy.DecisionUnknown {
		if !unknownMatches(observed.Unknown, reason, incompatible > 0) {
			return result, fmt.Errorf("fixture %q unknown state is not the exact six-field boundary", fixture.ID)
		}
	} else if observed.Unknown != nil {
		return result, fmt.Errorf("fixture %q non-unknown decision carries unknown state", fixture.ID)
	}
	if observed.ContractDigest != contractDigest || observed.EvaluatorDigest != evaluatorDigest || observed.PolicyActivitiesExpected != 1 || observed.PolicyActivitiesObserved != 1 || observed.GeneratedEvaluatorMismatches != 0 || observed.SemanticOperations != 1 || observed.RepositoryWrites != 0 || observed.LocalBuildExecutions != 0 || observed.LocalTestExecutions != 0 {
		return result, fmt.Errorf("fixture %q report binding or effect metrics are invalid", fixture.ID)
	}
	if observed.InvocationsObserved != len(entries) || observed.ComparableObservations != len(group) || observed.IncompatibleObservations != incompatible || observed.ContradictoryObservations != contradictory || observed.LedgerLines != countLines(rawLedger) || observed.LedgerBytes != len(rawLedger) {
		return result, fmt.Errorf("fixture %q report counts differ from raw ledger", fixture.ID)
	}
	if receipt.Schema != publicdiscovery.InvocationReceiptSchema || receipt.Sequence != current.Sequence || receipt.ObservationDigest != current.ObservationDigest || receipt.GroupKeyDigest != keyDigest || receipt.Decision != observed.Decision || receipt.Reason != observed.Reason || receipt.LedgerLines != observed.LedgerLines || receipt.LedgerBytes != observed.LedgerBytes || receipt.ArtifactDenominator != observed.ArtifactDenominator || receipt.ArtifactCount != observed.ArtifactCount || receipt.CandidateEmitted != (decision == discoverypolicy.DecisionClosed) {
		return result, fmt.Errorf("fixture %q receipt differs from report or ledger", fixture.ID)
	}
	wantArtifacts := 4
	if decision == discoverypolicy.DecisionClosed {
		wantArtifacts = 5
	}
	if observed.ArtifactDenominator != wantArtifacts || observed.ArtifactCount != wantArtifacts || observed.CandidatesEmitted != boolInt(decision == discoverypolicy.DecisionClosed) {
		return result, fmt.Errorf("fixture %q artifact boundary is invalid", fixture.ID)
	}
	if decision == discoverypolicy.DecisionClosed {
		candidateBytes, err := independentCandidate(current, keyDigest, reason)
		if err != nil {
			return result, err
		}
		candidatePath := observed.CandidatePath
		if candidatePath == "" {
			candidatePath = fixture.Candidate
		}
		candidateData, readErr := os.ReadFile(candidatePath)
		if readErr != nil {
			return result, fmt.Errorf("fixture %q candidate: %w", fixture.ID, readErr)
		}
		result.CandidateBytesEqual = bytes.Equal(candidateBytes, candidateData)
		replayBytes, replayErr := independentCandidate(current, keyDigest, reason)
		if replayErr != nil {
			return result, replayErr
		}
		result.CandidateReplayEqual = bytes.Equal(candidateBytes, replayBytes)
		if !result.CandidateBytesEqual || !result.CandidateReplayEqual || !observed.CandidateBytesEqual || !observed.CandidateByteReplayEqual {
			return result, fmt.Errorf("fixture %q closed candidate bytes are not deterministic: file_equal=%t replay_equal=%t report_equal=%t report_replay_equal=%t path=%q", fixture.ID, result.CandidateBytesEqual, result.CandidateReplayEqual, observed.CandidateBytesEqual, observed.CandidateByteReplayEqual, candidatePath)
		}
	} else if observed.CandidatePath != "" || observed.CandidateBytesEqual || observed.CandidateByteReplayEqual || observed.CandidateByteMismatches != 0 {
		return result, fmt.Errorf("fixture %q non-closed report emitted candidate evidence", fixture.ID)
	}
	return result, nil
}

func unknownMatches(unknown *generation.EnvelopeUnknownState, reason string, incompatible bool) bool {
	if unknown == nil || unknown.Stage != "PUBLIC_SELF_OBSERVATION" || unknown.Step != "COMPARE_CALLER_OWNED_LEDGER" || unknown.Reason != reason || unknown.UnknownClass != "INCOMPLETE_EVIDENCE" || unknown.NextOperation != "RECORD_ONE_MORE_EXACT_COMPARABLE_GENERATE" {
		return false
	}
	wantBlockedBy := []string{"comparable_observation_quorum"}
	if incompatible {
		wantBlockedBy = []string{"same_source_input_toolchain_contract_group"}
	}
	return sameStrings(unknown.BlockedBy, wantBlockedBy)
}

func independentlyParseLedger(raw []byte) ([]publicdiscovery.LedgerEntry, error) {
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return nil, errors.New("ledger is empty or missing terminal newline")
	}
	lines := bytes.Split(raw, []byte{'\n'})
	lines = lines[:len(lines)-1]
	entries := make([]publicdiscovery.LedgerEntry, 0, len(lines))
	for index, line := range lines {
		var entry publicdiscovery.LedgerEntry
		if err := decodeStrict(line, &entry); err != nil {
			return nil, fmt.Errorf("ledger line %d: %w", index+1, err)
		}
		if entry.Sequence != index+1 || entry.Schema != publicdiscovery.ObservationLedgerSchema || entry.Operation != publicdiscovery.Operation || entry.RepositoryWrites != 0 || entry.LocalBuildExecutions != 0 || entry.LocalTestExecutions != 0 {
			return nil, fmt.Errorf("ledger line %d has invalid sequence, schema, or effect", index+1)
		}
		if !allKnownDigests(entry) {
			return nil, fmt.Errorf("ledger line %d has an unknown digest", index+1)
		}
		want, err := publicdiscovery.EntryContentDigest(entry)
		if err != nil || want != entry.ObservationDigest {
			return nil, fmt.Errorf("ledger line %d content digest is invalid", index+1)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func forgeContradiction(inputPath, outputPath string) error {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read forge input: %w", err)
	}
	entries, err := independentlyParseLedger(raw)
	if err != nil {
		return fmt.Errorf("forge input ledger: %w", err)
	}
	if len(entries) < 2 {
		return errors.New("forge contradiction requires at least two valid ledger entries")
	}
	entries[len(entries)-1].GeneratedOutputDigest = cache.HashBytes([]byte("public-discovery-forged-contradiction")).String()
	entries[len(entries)-1].ObservationDigest, err = publicdiscovery.EntryContentDigest(entries[len(entries)-1])
	if err != nil {
		return err
	}
	var output bytes.Buffer
	for _, entry := range entries {
		data, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			return marshalErr
		}
		output.Write(data)
		output.WriteByte('\n')
	}
	if err := os.WriteFile(outputPath, output.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write forged ledger: %w", err)
	}
	return nil
}

func expectedClassification(comparable, incompatible, contradictory int) (string, string, []string) {
	switch {
	case contradictory > 0:
		return discoverypolicy.DecisionRefuted, publicdiscovery.ReasonContradictory, []string{discoverypolicy.CaseContradictory}
	case comparable < publicdiscovery.Quorum && incompatible > 0:
		return discoverypolicy.DecisionUnknown, publicdiscovery.ReasonIncompatible, []string{discoverypolicy.CaseIncompatibleGroup}
	case comparable < publicdiscovery.Quorum:
		return discoverypolicy.DecisionUnknown, publicdiscovery.ReasonSingle, []string{discoverypolicy.CaseSingleComparable}
	default:
		return discoverypolicy.DecisionClosed, publicdiscovery.ReasonClosed, []string{discoverypolicy.CaseExactPairCandidate, discoverypolicy.CaseDeterministicReplay}
	}
}

func expectedReason(caseID string) string {
	switch caseID {
	case discoverypolicy.CaseExactPairCandidate, discoverypolicy.CaseDeterministicReplay:
		return publicdiscovery.ReasonClosed
	case discoverypolicy.CaseSingleComparable:
		return publicdiscovery.ReasonSingle
	case discoverypolicy.CaseIncompatibleGroup:
		return publicdiscovery.ReasonIncompatible
	case discoverypolicy.CaseTamperedEntry:
		return publicdiscovery.ReasonTampered
	case discoverypolicy.CaseContradictory:
		return publicdiscovery.ReasonContradictory
	default:
		return ""
	}
}

func contradictoryCount(entries []publicdiscovery.LedgerEntry) int {
	if len(entries) < 2 {
		return 0
	}
	first := []string{entries[0].GeneratedSemanticDigest, entries[0].GeneratedOutputDigest, entries[0].GeneratedManifestDigest}
	count := 0
	for _, entry := range entries[1:] {
		current := []string{entry.GeneratedSemanticDigest, entry.GeneratedOutputDigest, entry.GeneratedManifestDigest}
		if !sameStrings(current, first) {
			count++
		}
	}
	return count
}

func independentCandidate(entry publicdiscovery.LedgerEntry, keyDigest, reason string) ([]byte, error) {
	candidate := publicdiscovery.Candidate{Schema: publicdiscovery.CandidateSchema,
		CandidateID: "public-generation-discovery/candidate/" + keyDigest, Operation: publicdiscovery.Operation,
		Decision: discoverypolicy.DecisionClosed, Reason: reason, GroupKeyDigest: keyDigest,
		SourceDigest: entry.SourceDigest, InputSemanticDigest: entry.InputSemanticDigest, PreviousGoDigest: entry.PreviousGoDigest,
		ToolchainDigest: entry.ToolchainDigest, ContractDigest: entry.ContractDigest, EvaluatorDigest: entry.EvaluatorDigest,
		GeneratedSemanticDigest: entry.GeneratedSemanticDigest, GeneratedOutputDigest: entry.GeneratedOutputDigest,
		GeneratedManifestDigest: entry.GeneratedManifestDigest, Quorum: publicdiscovery.Quorum,
		ExecutionAllowed: false, AuthorizationRequired: true, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0}
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func allKnownDigests(entry publicdiscovery.LedgerEntry) bool {
	values := []string{entry.SourceDigest, entry.InputSemanticDigest, entry.PreviousGoDigest, entry.ToolchainDigest,
		entry.ContractDigest, entry.EvaluatorDigest, entry.GeneratedSemanticDigest, entry.GeneratedOutputDigest, entry.GeneratedManifestDigest, entry.ObservationDigest}
	for _, value := range values {
		if !cache.Digest(value).Known() {
			return false
		}
	}
	return true
}

func countLines(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	return bytes.Count(raw, []byte{'\n'})
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func exitError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
