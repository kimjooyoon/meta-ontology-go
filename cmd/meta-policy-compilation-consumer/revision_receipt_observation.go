package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// Receipt interpretation uses consumer-owned wire declarations and raw Gooo.
// Producer execution remains a separate, unobserved claim in this observer.
const revisionReceiptSchema = "gooo/meta-policy-revision-receipt-observation/v1"

type revisionReceiptCheck struct {
	ID      string                    `json:"id"`
	State   string                    `json:"state"`
	Reason  string                    `json:"reason"`
	Unknown *sourceObservationUnknown `json:"unknown,omitempty"`
}

type revisionReceiptObservation struct {
	Schema                 string                   `json:"schema"`
	SourceDigest           string                   `json:"source_digest"`
	RequestArtifactDigest  string                   `json:"request_artifact_digest"`
	CanonicalRequestDigest string                   `json:"canonical_request_digest"`
	ReportArtifactDigest   string                   `json:"report_artifact_digest"`
	Decision               string                   `json:"decision"`
	Checks                 []revisionReceiptCheck   `json:"checks"`
	Counts                 revisionWireCounts       `json:"reconstructed_counts"`
	Transitions            []revisionWireTransition `json:"reconstructed_transitions"`
	BaselineSource         policySourceObservation  `json:"baseline_source"`
	CandidateSource        policySourceObservation  `json:"candidate_source"`
	IndependentComparisons int                      `json:"independent_result_comparisons"`
	ExecutionObserved      bool                     `json:"policy_execution_observed"`
	ExecutionUnknown       sourceObservationUnknown `json:"execution_unknown"`
	SharedAssumptions      []string                 `json:"shared_assumptions"`
	UnobservedClaims       []string                 `json:"unobserved_claims"`
	Improvement            string                   `json:"improvement"`
	MutationAuthority      int                      `json:"mutation_authority"`
	PromotionAuthority     int                      `json:"promotion_authority"`
}

type revisionReceiptRule struct {
	Condition     string   `json:"condition"`
	Decision      string   `json:"decision"`
	Stage         string   `json:"stage"`
	Step          int      `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func revisionReceiptMode(flags *flag.FlagSet) (bool, error) {
	requested := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "observe-revision-receipt" || f.Name == "revision-request" {
			requested = true
		}
	})
	if !requested {
		return false, nil
	}
	var modeError error
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "observe-revision-receipt", "revision-request", "policy", "profile-package", "profile-namespace":
		default:
			modeError = fmt.Errorf("revision receipt observation does not accept -%s", f.Name)
		}
	})
	for _, name := range []string{"policy", "revision-request", "observe-revision-receipt", "profile-package", "profile-namespace"} {
		f := flags.Lookup(name)
		if f == nil || strings.TrimSpace(f.Value.String()) == "" {
			modeError = fmt.Errorf("revision receipt observation requires -%s", name)
		}
	}
	if flags.NArg() != 0 {
		modeError = errors.New("revision receipt observation does not accept positional arguments")
	}
	return true, modeError
}

func runRevisionReceiptObservation(policyPath, requestPath, reportPath, pkg, namespace string, output io.Writer) error {
	inputs := make([][]byte, 0, 3)
	for _, name := range []string{policyPath, requestPath, reportPath} {
		file, err := os.Open(name)
		if err != nil {
			return fmt.Errorf("open revision receipt input: %w", err)
		}
		data, readError := io.ReadAll(io.LimitReader(file, (16<<20)+1))
		closeError := file.Close()
		if err := errors.Join(readError, closeError); err != nil {
			return fmt.Errorf("read revision receipt input: %w", err)
		}
		if len(data) > 16<<20 {
			return errors.New("revision receipt input exceeds 16 MiB")
		}
		inputs = append(inputs, data)
	}
	report, observationError := reconstructRevisionReceipt(policyPath, inputs[0], inputs[1], inputs[2], pkg, namespace)
	if report.Schema == "" {
		return observationError
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return errors.Join(observationError, encoder.Encode(report))
}

func decodeRevisionReceiptDocument(data []byte, target any, required []string) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var scan func(int) error
	scan = func(depth int) error {
		if depth > 128 {
			return errors.New("revision receipt JSON nesting exceeds 128")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token {
		case json.Delim('{'):
			seen := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok || seen[key] {
					return fmt.Errorf("duplicate or invalid revision receipt JSON key %v", keyToken)
				}
				seen[key] = true
				if err := scan(depth + 1); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
		case json.Delim('['):
			for decoder.More() {
				if err := scan(depth + 1); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
		}
		return err
	}
	if err := scan(0); err != nil {
		return err
	}
	return decodeRequiredDocument(data, target, required)
}

func revisionReceiptEqual(left, right any) bool {
	leftBytes, leftError := json.Marshal(left)
	rightBytes, rightError := json.Marshal(right)
	if leftError != nil || rightError != nil {
		return false
	}
	var a, b any
	if json.Unmarshal(leftBytes, &a) != nil || json.Unmarshal(rightBytes, &b) != nil {
		return false
	}
	return reflect.DeepEqual(a, b)
}

func revisionReceiptRules(source policySourceObservation) ([]revisionReceiptRule, error) {
	data, err := json.Marshal(source.DecisionRules)
	if err != nil {
		return nil, err
	}
	var rows []revisionReceiptRule
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func revisionReceiptPolicyMatches(observed policySourceObservation, reported revisionWirePolicy) bool {
	data, err := json.Marshal(reported)
	if err != nil {
		return false
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return false
	}
	bindings := map[string]any{
		"schema": "gooo/meta-policy-compilation/v3", "policy_id": observed.PolicyID,
		"package": observed.Package, "namespace": observed.Namespace,
		"source_digest": observed.SourceDigest, "semantic_digest": observed.SemanticDigest,
		"fixed_denominator": observed.Denominator, "rules": observed.RuleBindings,
		"decision_reduction": map[string]any{"schema": "decision-reduction:v2", "rules": observed.DecisionRules},
	}
	for key, want := range bindings {
		var got any
		if json.Unmarshal(fields[key], &got) != nil || !revisionReceiptEqual(got, want) {
			return false
		}
	}
	return true
}

func revisionReceiptKnownDecision(value string) bool {
	return value == "PASS" || value == "FAIL_CLOSED" || value == "UNKNOWN"
}

var revisionReceiptDigest = regexp.MustCompile("^sha256:[0-9a-f]{64}$")

func revisionReceiptEvaluate(source policySourceObservation, rows []revisionReceiptRule, input revisionWireCase) revisionWireResult {
	sourceOK := revisionReceiptDigest.MatchString(input.ObservedSourceDigest)
	artifactOK := revisionReceiptDigest.MatchString(input.ObservedArtifactSourceDigest)
	independentOK := revisionReceiptDigest.MatchString(input.ObservedIndependentDigest)
	judgeOK := revisionReceiptDigest.MatchString(input.ObservedGeneratedJudgeDigest)
	ready := input.ProducerAvailable && input.ConsumerAvailable
	empty := input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" ||
		input.ObservedIndependentDigest == "" || input.ObservedGeneratedJudgeDigest == ""
	matches := map[string]bool{
		"UNRECOGNIZED_TOP_LEVEL_DECISION": input.UpperDecision != "" && !revisionReceiptKnownDecision(input.UpperDecision),
		"SOURCE_DIGEST_MISMATCH":          sourceOK && input.ObservedSourceDigest != source.SourceDigest,
		"ARTIFACT_SOURCE_MISMATCH":        sourceOK && artifactOK && input.ObservedSourceDigest == source.SourceDigest && input.ObservedArtifactSourceDigest != source.SourceDigest,
		"INDEPENDENT_SOURCE_MISMATCH": sourceOK && artifactOK && independentOK && input.ObservedSourceDigest == source.SourceDigest &&
			input.ObservedArtifactSourceDigest == source.SourceDigest && input.ObservedIndependentDigest != source.SemanticDigest,
		"EVIDENCE_UNAVAILABLE": !ready && !sourceOK && !artifactOK && !independentOK && !judgeOK,
		"DIGEST_UNAVAILABLE":   ready && empty,
		"MALFORMED_DIGEST":     ready && !empty && !(sourceOK && artifactOK && independentOK && judgeOK),
		"SEMANTIC_EQUIVALENCE": ready && sourceOK && artifactOK && independentOK && judgeOK &&
			input.ObservedSourceDigest == source.SourceDigest && input.ObservedArtifactSourceDigest == source.SourceDigest &&
			input.ObservedIndependentDigest == source.SemanticDigest,
	}
	result := revisionWireResult{
		CaseID: input.ID, PolicyDigest: source.SourceDigest, SemanticDigest: source.SemanticDigest,
		Denominator: source.Denominator, Decision: "FAIL_CLOSED", Stage: "COMPILE",
		Reason: "NO_REDUCTION_RULE_MATCHED", BlockedBy: []string{},
	}
	for _, row := range rows {
		if !matches[row.Condition] {
			continue
		}
		result.Decision, result.MatchedCondition = row.Decision, row.Condition
		result.Stage, result.Step, result.Reason = row.Stage, row.Step, row.Reason
		if row.Decision == "UNKNOWN" {
			result.UnknownClass, result.NextOperation = row.UnknownClass, row.NextOperation
			result.BlockedBy = append([]string(nil), row.BlockedBy...)
		}
		break
	}
	return result
}

func reconstructRevisionReceipt(filename string, source, requestBytes, reportBytes []byte, pkg, namespace string) (revisionReceiptObservation, error) {
	var request revisionWireRequest
	if err := decodeRevisionReceiptDocument(requestBytes, &request, []string{
		"expected_source_digest", "condition", "from_decision", "to_decision", "cases",
	}); err != nil {
		return revisionReceiptObservation{}, fmt.Errorf("decode original revision request: %w", err)
	}
	var supplied revisionWireObservation
	if err := decodeRevisionReceiptDocument(reportBytes, &supplied, []string{
		"schema", "source_file", "canonical_request_digest", "request_artifact_digest", "request",
		"original_policy", "candidate_policy", "candidate_source", "changed_coordinates",
		"baseline", "candidate", "transitions", "counts", "execution_status", "execution_conformance",
		"admission", "pending", "input_provenance", "repository_observation", "improvement",
		"mutation_authority", "promotion_authority",
	}); err != nil {
		return revisionReceiptObservation{}, fmt.Errorf("decode supplied revision report: %w", err)
	}
	canonical, err := json.Marshal(request)
	if err != nil {
		return revisionReceiptObservation{}, err
	}
	observed := revisionReceiptObservation{
		Schema: revisionReceiptSchema, SourceDigest: digestBytes(source),
		RequestArtifactDigest: digestBytes(requestBytes), CanonicalRequestDigest: digestBytes(canonical),
		ReportArtifactDigest: digestBytes(reportBytes), Decision: "RECEIPT_CONSISTENT_ONLY",
		Checks: []revisionReceiptCheck{}, Transitions: []revisionWireTransition{},
		ExecutionUnknown: sourceObservationUnknown{
			State: "UNKNOWN", Stage: "EXECUTION_PROVENANCE", Step: "OBSERVE_REVISION_RECEIPT",
			Reason: "PROCESS_EXECUTION_NOT_ATTESTED", UnknownClass: "DIRECT_MISSING",
			NextOperation: "BIND_INDEPENDENT_NATIVE_PROCESS_EVIDENCE", BlockedBy: []string{},
		},
		SharedAssumptions: []string{"GOOO_SYNTAX_FRONTEND", "POLICY_COMPILATION_WIRE_SCHEMA"},
		UnobservedClaims: []string{"PROCESS_EXECUTION", "GENERATED_PROGRAM_SEMANTICS", "POLICY_STRUCTURE_METRICS",
			"WALL_TIME_ACCURACY", "REPOSITORY_WRITE_SET", "EXTERNAL_UTILITY", "CAUSAL_IMPROVEMENT"},
		Improvement: "UNKNOWN",
	}
	check := func(id string, ok bool, reason string) {
		state := "CLOSED"
		if !ok {
			state, observed.Decision = "REFUTED", "REFUTED"
		}
		observed.Checks = append(observed.Checks, revisionReceiptCheck{ID: id, State: state, Reason: reason})
	}
	baseline, beforeError := reconstructSourceObservation(filename, source, pkg, namespace)
	candidate, afterError := reconstructSourceObservation(filename, []byte(supplied.CandidateSource), pkg, namespace)
	if err := errors.Join(beforeError, afterError); err != nil {
		check("SOURCE_BINDING", false, "SOURCE_RECONSTRUCTION_FAILED")
		for _, id := range []string{"REQUEST_BINDING", "REVISION_SCOPE", "DECLARED_INPUT_PRESERVATION", "RESULT_RECONSTRUCTION", "ACCOUNTING_RECONSTRUCTION"} {
			unknown := sourceObservationUnknown{
				State: "UNKNOWN", Stage: "RECEIPT_RECONSTRUCTION", Step: id, Reason: "SOURCE_RECONSTRUCTION_REQUIRED",
				UnknownClass: "DEPENDENCY_BLOCKED", NextOperation: "REPAIR_SOURCE_BINDING", BlockedBy: []string{"SOURCE_BINDING"},
			}
			observed.Checks = append(observed.Checks, revisionReceiptCheck{ID: id, State: "UNKNOWN", Reason: unknown.Reason, Unknown: &unknown})
		}
		return observed, err
	}
	observed.BaselineSource, observed.CandidateSource = baseline, candidate
	check("SOURCE_BINDING", supplied.Schema == revisionWireSchema &&
		revisionReceiptPolicyMatches(baseline, supplied.OriginalPolicy) && revisionReceiptPolicyMatches(candidate, supplied.CandidatePolicy),
		"COMPARE_RAW_SOURCE_POLICY_PROJECTIONS")
	requestOK := request.ExpectedSourceDigest == baseline.SourceDigest && len(request.Cases) > 0 &&
		revisionReceiptKnownDecision(request.FromDecision) && revisionReceiptKnownDecision(request.ToDecision) &&
		request.FromDecision != request.ToDecision && revisionReceiptEqual(request, supplied.Request) &&
		observed.CanonicalRequestDigest == supplied.RequestDigest && observed.RequestArtifactDigest == supplied.RequestArtifactDigest
	seen := map[string]bool{}
	for _, pair := range request.Cases {
		id := pair.Baseline.ID
		requestOK = requestOK && strings.TrimSpace(id) != "" && id == pair.Candidate.ID && !seen[id] &&
			strings.TrimSpace(pair.Baseline.EvidenceClass) != "" && strings.TrimSpace(pair.Candidate.EvidenceClass) != "" &&
			strings.TrimSpace(pair.Baseline.Provenance) != "" && strings.TrimSpace(pair.Candidate.Provenance) != ""
		seen[id] = true
	}
	check("REQUEST_BINDING", requestOK, "COMPARE_ORIGINAL_RAW_AND_CANONICAL_REQUEST")
	beforeRules, beforeError := revisionReceiptRules(baseline)
	afterRules, afterError := revisionReceiptRules(candidate)
	if err := errors.Join(beforeError, afterError); err != nil {
		return observed, err
	}
	scopeOK := baseline.PolicyID == candidate.PolicyID && baseline.Package == candidate.Package &&
		baseline.Namespace == candidate.Namespace && baseline.Denominator == candidate.Denominator &&
		baseline.SourceDigest != candidate.SourceDigest && revisionReceiptEqual(baseline.RuleBindings, candidate.RuleBindings) &&
		len(beforeRules) == len(afterRules) && revisionReceiptEqual(supplied.ChangedCoordinates, []string{"transition.to", "case.resolution.decision"})
	changes := 0
	for index, before := range beforeRules {
		if index >= len(afterRules) {
			scopeOK = false
			break
		}
		after := afterRules[index]
		if before.Condition == request.Condition {
			changes++
			scopeOK = scopeOK && before.Decision == request.FromDecision && after.Decision == request.ToDecision
			after.Decision = before.Decision
		}
		scopeOK = scopeOK && revisionReceiptEqual(before, after)
	}
	check("REVISION_SCOPE", scopeOK && changes == 1, "COMPARE_ONLY_REQUESTED_TRANSITION_AND_RESOLUTION")
	inputsOK, resultsOK, incomplete := true, true, false
	counts := revisionWireCounts{RequestedCasePairs: len(request.Cases)}
	for side, phase := range []revisionWireExecution{supplied.Baseline, supplied.Candidate} {
		view, rows := baseline, beforeRules
		if side == 1 {
			view, rows = candidate, afterRules
		}
		attempted := side == 0 || supplied.Baseline.Complete
		if !attempted {
			inputsOK = inputsOK && reflect.DeepEqual(phase, revisionWireExecution{})
			incomplete = true
			continue
		}
		inputs := make([]revisionWireCase, 0, len(request.Cases))
		expected := make([]revisionWireResult, 0, len(request.Cases))
		for _, pair := range request.Cases {
			input := pair.Baseline
			if side == 1 {
				input = pair.Candidate
			}
			inputs = append(inputs, input)
			expected = append(expected, revisionReceiptEvaluate(view, rows, input))
		}
		inputsOK = inputsOK && revisionReceiptEqual(phase.DeclaredInputs, inputs)
		resultsOK = resultsOK && phase.GeneratedJudgeSource != "" &&
			digestBytes([]byte(phase.GeneratedJudgeSource)) == phase.GeneratedJudgeDigest && phase.WallMilliseconds >= 0 &&
			len(phase.SourceResults) == len(expected) && len(phase.FirstResults) <= len(expected) &&
			len(phase.ReplayResults) <= len(expected) &&
			(len(phase.ReplayResults) == 0 || len(phase.FirstResults) == len(expected))
		for index, result := range phase.SourceResults {
			observed.IndependentComparisons++
			resultsOK = resultsOK && index < len(expected) && revisionReceiptEqual(result, expected[index])
		}
		for _, returned := range [][]revisionWireResult{phase.FirstResults, phase.ReplayResults} {
			for index, result := range returned {
				counts.SourceComparisons++
				observed.IndependentComparisons++
				if index >= len(expected) || !revisionReceiptEqual(result, expected[index]) {
					counts.SourceMismatches++
					resultsOK = false
				}
			}
		}
		for index, result := range phase.ReplayResults {
			if index < len(phase.FirstResults) {
				counts.ReplayComparisons++
				if !revisionReceiptEqual(result, phase.FirstResults[index]) {
					counts.ReplayMismatches++
				}
			}
		}
		complete := phase.ExecutionError == "" && len(phase.FirstResults) == len(expected) && len(phase.ReplayResults) == len(expected)
		resultsOK = resultsOK && phase.Complete == complete
		incomplete = incomplete || !complete
		if phase.ExecutionError != "" {
			counts.FailedBatches++
		}
	}
	check("DECLARED_INPUT_PRESERVATION", inputsOK, "COMPARE_PAIRED_CALLER_SNAPSHOTS_WITHOUT_REBINDING")
	check("RESULT_RECONSTRUCTION", resultsOK, "COMPARE_ALL_RESULT_FIELDS_WITH_CONSUMER_INTERPRETATION")
	if resultsOK && incomplete {
		last := &observed.Checks[len(observed.Checks)-1]
		last.State, last.Reason = "UNKNOWN", "EXECUTION_COHORT_INCOMPLETE"
		last.Unknown = &sourceObservationUnknown{
			State: "UNKNOWN", Stage: "RECEIPT_RECONSTRUCTION", Step: "RESULT_RECONSTRUCTION",
			Reason: last.Reason, UnknownClass: "DIRECT_MISSING",
			NextOperation: "SUPPLY_MISSING_EXECUTION_RESULTS", BlockedBy: []string{},
		}
		if observed.Decision != "REFUTED" {
			observed.Decision = "UNKNOWN"
		}
	}
	for index, pair := range request.Cases {
		if index >= len(supplied.Baseline.FirstResults) || index >= len(supplied.Candidate.FirstResults) {
			continue
		}
		before, after := supplied.Baseline.FirstResults[index], supplied.Candidate.FirstResults[index]
		transition := revisionWireTransition{
			CaseID: pair.Baseline.ID, InputsIdentical: pair.Baseline == pair.Candidate,
			BaselineCondition: before.MatchedCondition, CandidateCondition: after.MatchedCondition,
			BaselineDecision: before.Decision, CandidateDecision: after.Decision,
			RequestedTransitionObserved: before.MatchedCondition == request.Condition && after.MatchedCondition == request.Condition &&
				before.Decision == request.FromDecision && after.Decision == request.ToDecision,
			CausalAttribution: "UNASSESSED",
		}
		counts.ObservedCasePairs++
		if transition.InputsIdentical {
			counts.SameInputObservedPairs++
		}
		if transition.RequestedTransitionObserved {
			counts.RequestedTransitionsObserved++
		}
		observed.Transitions = append(observed.Transitions, transition)
	}
	observed.Counts = counts
	status, conformance := "NOT_COMPLETED", "UNKNOWN"
	if counts.FailedBatches > 0 {
		status = "FAILED"
	} else if supplied.Baseline.Complete && supplied.Candidate.Complete {
		status, conformance = "COMPLETED", "PASS"
	}
	if counts.SourceMismatches > 0 || counts.ReplayMismatches > 0 {
		conformance = "REFUTED"
	}
	pending := []revisionWirePending{}
	pendingRecord := func(stage, step, reason, next string) revisionWirePending {
		return revisionWirePending{
			State: "UNKNOWN", Stage: stage, Step: step, Reason: reason,
			UnknownClass: "DIRECT_MISSING", NextOperation: next, BlockedBy: []string{},
		}
	}
	for side, phase := range []revisionWireExecution{supplied.Baseline, supplied.Candidate} {
		if phase.ExecutionError != "" {
			name := "BASELINE"
			if side == 1 {
				name = "CANDIDATE"
			}
			pending = append(pending, pendingRecord(name+"_EXECUTION", "BUILD_AND_EXECUTE_GENERATED_BATCH",
				name+"_GENERATED_BATCH_FAILED", "RERUN_"+name+"_GENERATED_BATCH"))
		}
	}
	if counts.RequestedTransitionsObserved == 0 {
		pending = append(pending, pendingRecord("REQUEST_COVERAGE", "OBSERVE_REQUESTED_TRANSITION",
			"REQUESTED_TRANSITION_NOT_OBSERVED", "SUPPLY_CASE_PAIR_EXERCISING_REQUESTED_CONDITION"))
	}
	admission := pendingRecord("INDEPENDENT_VALIDATION", "OBSERVE_REVISION_CANDIDATE",
		"INDEPENDENT_REVISION_EVIDENCE_MISSING", "RUN_INDEPENDENT_REVISION_OBSERVER")
	check("ACCOUNTING_RECONSTRUCTION", supplied.Counts == counts && revisionReceiptEqual(supplied.Transitions, observed.Transitions) &&
		supplied.ExecutionStatus == status && supplied.ExecutionConformance == conformance &&
		revisionReceiptEqual(supplied.Pending, pending) && revisionReceiptEqual(supplied.Admission, admission) &&
		supplied.InputProvenance == "CALLER_DECLARED_NOT_VERIFIED" && supplied.RepositoryObservation == "NOT_PERFORMED" &&
		supplied.Improvement == "UNKNOWN" && supplied.MutationAuthority == 0 && supplied.PromotionAuthority == 0,
		"RECOMPUTE_ACCOUNTING_AND_PRESERVE_UNOBSERVED_AUTHORITY")
	if observed.Decision == "REFUTED" {
		return observed, errors.New("revision receipt contradicts its bound source, request or supplied records")
	}
	return observed, nil
}
