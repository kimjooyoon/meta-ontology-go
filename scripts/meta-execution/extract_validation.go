package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type extractValidation struct {
	BeforeFunctionLines  int
	AfterFunctionLines   int
	TransformedSubject   string
	AtomicReplacement    bool
	FormatFixedPoint     bool
	HeaderPreserved      bool
	ImportIdentity       bool
	PackageConformance   bool
	ProjectedTestsPassed bool
}

type extractValidationUnknown struct {
	reason string
}

func (err *extractValidationUnknown) Error() string {
	return err.reason
}

func validateExtractedFiles(root string, subject sourcepolicy.SourceSubject, before []byte, observed extractorSubject, replacements []namespaceReplacementReceipt, backup backupCleanupObservation) (extractValidation, error) {
	if err := validateBackupCleanup(backup, replacements); err != nil {
		return extractValidation{}, err
	}
	beforeSet, beforeFile, err := parseGoFile(subject.Path, before)
	if err != nil {
		return extractValidation{}, err
	}
	beforePackage := beforeFile.Name.Name
	beforeImports := importIdentity(beforeFile)
	beforeHeader := sourceHeader(before, beforeSet, beforeFile)
	beforeFunction, exists := functionInFile(beforeFile, subject.Path, subject)
	if !exists {
		return extractValidation{}, fmt.Errorf("function %s is absent before extraction", subject.Name)
	}
	found, afterLines, seen, err := validateOutputFiles(root, subject, beforePackage, beforeImports, beforeHeader, observed)
	if err != nil {
		return extractValidation{}, err
	}
	if found != 1 || !seen[subject.Path] {
		return extractValidation{}, fmt.Errorf("function %s found in %d outputs", subject.Name, found)
	}
	atomicReplacement, err := validateNamespaceReplacements(root, observed, replacements)
	if err != nil {
		return extractValidation{}, err
	}
	return extractValidation{
		BeforeFunctionLines: declarationLinesFor(beforeSet, beforeFunction),
		AfterFunctionLines:  afterLines,
		TransformedSubject:  subject.Path + "#" + subject.Name + "=>" + strings.Join(sortedKeys(seen), ","),
		AtomicReplacement:   atomicReplacement,
		FormatFixedPoint:    true,
		HeaderPreserved:     true,
		ImportIdentity:      true,
		PackageConformance:  true,
	}, nil
}

func decodeExtractorReport(path, expectedSHA string) ([]byte, extractorReport, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, extractorReport{}, err
	}
	var report extractorReport
	var presence extractorReportPresence
	if err := decodeStrictBytes(raw, &report); err != nil ||
		decodeStrictBytes(raw, &presence) != nil ||
		report.Schema != functionExtractionReportSchema || report.SourceSHA != expectedSHA ||
		!validStrategyEvidenceReport(report, presence) {
		return raw, extractorReport{}, fmt.Errorf("malformed extraction report")
	}
	return raw, report, nil
}

type extractorReportPresence struct {
	Schema                string                        `json:"schema"`
	SourceSHA             string                        `json:"source_sha"`
	StagedSubjects        int                           `json:"staged_subjects"`
	Subjects              []extractorSubjectPresence    `json:"subjects"`
	Unhandled             []string                      `json:"unhandled"`
	Failures              []extractorFailureRecord      `json:"failures,omitempty"`
	Indicators            []json.RawMessage             `json:"indicators"`
	NamespaceReplacements []namespaceReplacementReceipt `json:"namespace_replacements,omitempty"`
	BackupCleanup         backupCleanupObservation      `json:"backup_cleanup"`
}

type extractorSubjectPresence struct {
	Logical      string                      `json:"logical"`
	State        string                      `json:"state"`
	Before       int                         `json:"before_lines"`
	After        int                         `json:"after_lines"`
	Files        []string                    `json:"changed_files"`
	CreatedFiles []string                    `json:"created_files,omitempty"`
	Consumer     string                      `json:"consumer"`
	Operation    string                      `json:"meta_operation"`
	Operations   []string                    `json:"meta_operations,omitempty"`
	Proof        string                      `json:"proof_choice"`
	Evidence     presentStrategyEvidenceList `json:"strategy_evidence"`
}

type presentStrategyEvidenceList struct {
	Present bool
	Values  []strategyEvidencePresence
}

func (list *presentStrategyEvidenceList) UnmarshalJSON(data []byte) error {
	list.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		list.Values = nil
		return nil
	}
	return decodeStrictBytes(data, &list.Values)
}

type strategyEvidencePresence struct {
	Strategy                    string                    `json:"strategy"`
	Operation                   string                    `json:"operation"`
	ContractActivity            string                    `json:"contract_activity"`
	ContractInputEntity         string                    `json:"contract_input_entity"`
	ContractOutputEntity        string                    `json:"contract_output_entity"`
	ContractInputSubjectKind    string                    `json:"contract_input_subject_kind"`
	ContractSourceDigest        string                    `json:"contract_source_digest"`
	ContractSemanticDigest      string                    `json:"contract_semantic_digest"`
	UsedInputFact               bool                      `json:"used_input_fact"`
	GeneratedOutputFact         bool                      `json:"generated_output_fact"`
	Subject                     string                    `json:"subject"`
	Helper                      string                    `json:"helper"`
	BeforeBytes                 int                       `json:"before_bytes"`
	AfterBytes                  int                       `json:"after_bytes"`
	BeforeFunctionLines         int                       `json:"before_function_lines"`
	AfterFunctionLines          int                       `json:"after_function_lines"`
	RenderedHelperBytes         int                       `json:"rendered_helper_bytes"`
	RenderedHelperLines         int                       `json:"rendered_helper_lines"`
	RenderedOuterHelperBytes    int                       `json:"rendered_outer_helper_bytes"`
	RenderedOuterHelperLines    int                       `json:"rendered_outer_helper_lines"`
	Obligations                 []projectionextractor.ObligationEvidence         `json:"obligations"`
	ContractObligations         []projectionextractor.ContractObligationEvidence `json:"contract_obligations"`
	ProofStages                 []proofStagePresence                             `json:"proof_stages"`
	FinalGeneratedBytes         int                       `json:"final_generated_bytes"`
	FinalGeneratedEvidenceBytes int                       `json:"final_generated_evidence_bytes"`
	FinalGeneratedUnits         int                       `json:"final_generated_units"`
}

type proofStagePresence struct {
	Name             string `json:"name"`
	Activity         string `json:"activity"`
	InputEntity      string `json:"input_entity"`
	OutputEntity     string `json:"output_entity"`
	Status           string `json:"status"`
	SourceDigest     string `json:"source_digest"`
	CandidateDigest  string `json:"candidate_digest"`
	InputEvidenceID  string `json:"input_evidence_id"`
	OutputEvidenceID string `json:"output_evidence_id"`
	PayloadDigest    string `json:"payload_digest"`
	PayloadBytes     *int   `json:"payload_bytes"`
	Detail           string `json:"detail,omitempty"`
}

func failedExtractionError(root, reportName string, plan generation.Plan, action generation.Action, observed ...generation.ProcessObservation) *operationError {
	path := filepath.Join(root, reportName)
	_, report, err := decodeExtractorReport(path, plan.HeadSHA)
	if err != nil {
		if os.IsNotExist(err) {
			return newOperationError("execute-operation", "run-function-extractor", "EXECUTOR_PROCESS_FAILED", "DIRECT_MISSING", "restore-operation-evidence")
		}
		return newOperationError("evaluate-operation", "decode-function-extraction-report", "INSTANCE_EVIDENCE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if failure := adjudicateExtractorReport(report); failure != nil {
		failure.counterexample = extractionCounterexample(action.Subject)
		failure.derivedRelations = extractorDerivedRelations(report, action)
		failure.evidence = extractionFailureEvidence(report, action)
		if len(observed) != 0 {
			failure.canonical = canonicalStructuredExtractorFailure(action.Subject, failure, observed[0])
		}
		return failure
	}
	if validationErr := validateBackupCleanup(report.BackupCleanup, report.NamespaceReplacements); validationErr != nil {
		reason := extractValidationErrorReason(validationErr)
		class := extractValidationErrorClass(validationErr)
		next := extractValidationNextOperation(validationErr)
		return newOperationError("evaluate-operation", "validate-function-extraction", reason, class, next)
	}
	return newOperationError("execute-operation", "run-function-extractor", "EXECUTOR_PROCESS_FAILED", "DIRECT_MISSING", "restore-operation-evidence")
}

func extractionCounterexample(subject string) string {
	parsed, err := sourcepolicy.ParseSourceSubject(subject)
	if err != nil {
		return ""
	}
	return parsed.Path + "#func:" + parsed.Name
}

func extractionFailureEvidence(report extractorReport, action generation.Action) []generation.ObservationFailureEvidence {
	counterexample := "extractor-report"
	if relations := extractorDerivedRelations(report, action); len(relations) != 0 {
		counterexample = relations[0].Counterexample
	}
	result := make([]generation.ObservationFailureEvidence, 0, len(action.RequiredIndicatorIDs))
	for _, identifier := range action.RequiredIndicatorIDs {
		result = append(result, generation.ObservationFailureEvidence{
			IndicatorID: identifier, Observed: 0, Expected: 1,
			Decision: "UNKNOWN", Counterexample: counterexample,
		})
	}
	return result
}

func extractorDerivedRelations(report extractorReport, action generation.Action) []generation.CounterexampleRelation {
	root := extractionCounterexample(action.Subject)
	seen := make(map[string]bool)
	result := make([]generation.CounterexampleRelation, 0, len(report.Failures))
	for _, failure := range report.Failures {
		if failure.Decision != "REFUTED" || failure.BlockerID == "" || seen[failure.BlockerID] {
			continue
		}
		seen[failure.BlockerID] = true
		result = append(result, generation.CounterexampleRelation{
			Counterexample: failure.BlockerID, DerivedFrom: root, Relation: "DERIVED_FROM",
		})
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Counterexample < result[right].Counterexample
	})
	return result
}

func canonicalStructuredExtractorFailure(subject string, failure *operationError, process generation.ProcessObservation) []byte {
	if failure == nil || failure.class != "KNOWN_CONTRADICTION" ||
		failure.stage != "derive-recipe" || failure.step != "select-declaration" ||
		failure.reason != "NO_SAFE_DECLARATION_CAPACITY" || len(failure.derivedRelations) != 1 {
		return nil
	}
	value := struct {
		ActionSubject  string `json:"action_subject"`
		Stage          string `json:"stage"`
		Step           string `json:"step"`
		Reason         string `json:"reason"`
		DerivedBlocker string `json:"derived_blocker"`
		ExitCode       int    `json:"exit_code"`
		StdoutBytes    int    `json:"stdout_bytes"`
		StderrBytes    int    `json:"stderr_bytes"`
	}{
		ActionSubject:  subject,
		Stage:          failure.stage,
		Step:           failure.step,
		Reason:         failure.reason,
		DerivedBlocker: failure.derivedRelations[0].Counterexample,
		ExitCode:       process.ExitCode,
		StdoutBytes:    process.StdoutBytes,
		StderrBytes:    process.StderrBytes,
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}

func adjudicateExtractorReport(report extractorReport) *operationError {
	if malformed := validateExtractorReport(report); malformed {
		return newOperationError("evaluate-operation", "decode-function-extraction-report", "INSTANCE_EVIDENCE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if failure, ok := reportFailureOperationError(report.Failures); ok {
		return failure
	}
	return nil
}

func validateExtractorReport(report extractorReport) bool {
	if !validStrategyEvidenceValues(report) {
		return true
	}
	failures := make(map[string]extractorFailureRecord, len(report.Failures))
	blockers := make(map[string]bool, len(report.Failures))
	for _, failure := range report.Failures {
		if !validExtractorFailure(failure) || failures[failure.Logical].Logical != "" {
			return true
		}
		if failure.BlockerID != "" && blockers[failure.BlockerID] {
			return true
		}
		failures[failure.Logical] = failure
		if failure.BlockerID != "" {
			blockers[failure.BlockerID] = true
		}
	}
	seenUnhandled := make(map[string]bool, len(report.Unhandled))
	for _, logical := range report.Unhandled {
		if logical == "" || seenUnhandled[logical] || failures[logical].Logical == "" {
			return true
		}
		seenUnhandled[logical] = true
	}
	if len(failures) != len(report.Unhandled) {
		return true
	}
	seenSubjects := make(map[string]bool, len(report.Subjects))
	seenChanged := make(map[string]bool)
	seenCreated := make(map[string]bool)
	for _, subject := range report.Subjects {
		if !validExtractionSubject(subject) || seenSubjects[subject.Logical] {
			return true
		}
		seenSubjects[subject.Logical] = true
		if _, exists := seenUnhandled[subject.Logical]; exists {
			return true
		}
		if !recordExtractionSubjectFiles(subject, seenChanged, seenCreated) {
			return true
		}
	}
	values, ok := extractorIndicatorValues(report.Indicators)
	if !ok || values["extraction.observed"] != len(report.Subjects)+len(report.Unhandled) ||
		values["extraction.staged"] != report.StagedSubjects || report.StagedSubjects != len(report.Subjects) ||
		values["extraction.unhandled"] != len(report.Unhandled) {
		return true
	}
	failed := len(failures) != 0 || len(report.Unhandled) != 0
	if failed {
		if values["extraction.applied"] != 0 || values["extraction.created"] != 0 {
			return true
		}
		for _, subject := range report.Subjects {
			if subject.State != "STAGED_NOT_COMMITTED" {
				return true
			}
		}
		return trueIfNoFailureBinding(report.Unhandled, failures)
	}
	if values["extraction.applied"] != len(report.Subjects) ||
		values["extraction.created"] != extractorCreatedCount(report.Subjects) {
		return true
	}
	for _, subject := range report.Subjects {
		if subject.State != "COMMITTED_APPLIED" {
			return true
		}
	}
	return false
}

func validStrategyEvidenceReport(report extractorReport, presence extractorReportPresence) bool {
	if len(report.Subjects) != len(presence.Subjects) || !validStrategyEvidenceValues(report) {
		return false
	}
	for index, subject := range report.Subjects {
		observed := presence.Subjects[index].Evidence
		if !observed.Present {
			if subject.Evidence != nil {
				return false
			}
			continue
		}
		if len(subject.Evidence) == 0 || len(subject.Evidence) != len(observed.Values) {
			return false
		}
		for evidenceIndex, evidence := range subject.Evidence {
			if len(evidence.ProofStages) != len(observed.Values[evidenceIndex].ProofStages) {
				return false
			}
			for stageIndex := range evidence.ProofStages {
				if observed.Values[evidenceIndex].ProofStages[stageIndex].PayloadBytes == nil {
					return false
				}
			}
		}
	}
	return true
}

func validStrategyEvidenceValues(report extractorReport) bool {
	for _, subject := range report.Subjects {
		for _, evidence := range subject.Evidence {
			if evidence.Strategy == "" || evidence.Operation == "" || evidence.ContractActivity == "" ||
				evidence.ContractInputEntity == "" || evidence.ContractOutputEntity == "" ||
				evidence.ContractInputSubjectKind == "" || evidence.ContractSourceDigest == "" ||
				evidence.ContractSemanticDigest == "" || !evidence.UsedInputFact || !evidence.GeneratedOutputFact ||
				evidence.Subject == "" || evidence.Helper == "" || evidence.BeforeBytes <= 0 ||
				evidence.AfterBytes <= 0 || evidence.BeforeFunctionLines <= 0 || evidence.AfterFunctionLines <= 0 ||
				evidence.RenderedHelperBytes <= 0 || evidence.RenderedHelperLines <= 0 ||
				evidence.RenderedOuterHelperBytes <= 0 || evidence.RenderedOuterHelperLines <= 0 ||
				evidence.FinalGeneratedBytes <= 0 || evidence.FinalGeneratedEvidenceBytes <= 0 ||
				evidence.FinalGeneratedUnits <= 0 || len(evidence.Obligations) == 0 ||
				len(evidence.ContractObligations) == 0 || len(evidence.ProofStages) == 0 {
				return false
			}
			for _, obligation := range evidence.Obligations {
				if obligation.Name == "" || obligation.Status == "" {
					return false
				}
			}
			for _, obligation := range evidence.ContractObligations {
				if obligation.Name == "" || obligation.Activity == "" || obligation.InputEntity == "" ||
					obligation.OutputEntity == "" || !obligation.UsedInputFact || !obligation.GeneratedOutputFact {
					return false
				}
			}
			for _, stage := range evidence.ProofStages {
				if stage.Name == "" || stage.Activity == "" || stage.InputEntity == "" ||
					stage.OutputEntity == "" || stage.Status == "" || stage.SourceDigest == "" ||
					stage.CandidateDigest == "" || stage.InputEvidenceID == "" || stage.OutputEvidenceID == "" ||
					stage.PayloadDigest == "" || stage.PayloadBytes < 0 {
					return false
				}
			}
		}
	}
	return true
}

func validExtractionSubject(subject extractorSubject) bool {
	if subject.After <= 0 {
		return false
	}
	if subject.Logical == "" || subject.Before <= 75 || subject.After > 75 || subject.After >= subject.Before ||
		(subject.State != "STAGED_NOT_COMMITTED" && subject.State != "COMMITTED_APPLIED") || subject.Consumer != "function-extractor" ||
		subject.Operation == "" || subject.Proof == "" || len(subject.Files) == 0 || len(subject.Operations) == 0 {
		return false
	}
	if subject.Proof != "axiomatic-foundation" && subject.Proof != "coherent-system" {
		return false
	}
	seen := make(map[string]bool, len(subject.Operations))
	found := false
	for _, operation := range subject.Operations {
		if operation == "" || seen[operation] {
			return false
		}
		seen[operation] = true
		if operation == subject.Operation {
			found = true
		}
	}
	return found
}

func recordExtractionSubjectFiles(subject extractorSubject, seenChanged, seenCreated map[string]bool) bool {
	files := make(map[string]bool, len(subject.Files))
	logicalCount := 0
	for _, path := range subject.Files {
		if path == "" || files[path] || seenChanged[path] {
			return false
		}
		files[path] = true
		seenChanged[path] = true
		if path == subject.Logical {
			logicalCount++
		}
	}
	if logicalCount != 1 {
		return false
	}
	created := make(map[string]bool, len(subject.CreatedFiles))
	for _, path := range subject.CreatedFiles {
		if path == "" || created[path] || !files[path] || seenCreated[path] {
			return false
		}
		created[path] = true
		seenCreated[path] = true
	}
	return true
}

func trueIfNoFailureBinding(unhandled []string, failures map[string]extractorFailureRecord) bool {
	for _, logical := range unhandled {
		if _, ok := failures[logical]; !ok {
			return true
		}
	}
	return false
}

func extractorIndicatorValues(raw []json.RawMessage) (map[string]int, bool) {
	expected := map[string]extractorIndicatorRecord{
		"extraction.observed":  {Limit: -1, Consumer: "function-extractor", Operation: "observe-density-residual", Proof: "axiomatic-foundation"},
		"extraction.staged":    {Limit: -1, Consumer: "function-extractor", Operation: "stage-helper-extraction", Proof: "coherent-system"},
		"extraction.applied":   {Limit: -1, Consumer: "logical-materializer", Operation: "accept-helper-extraction", Proof: "coherent-system"},
		"extraction.created":   {Limit: -1, Consumer: "authorized-write-set", Operation: "authorize-declared-file-creation", Proof: "axiomatic-foundation"},
		"extraction.unhandled": {Limit: 0, Blocking: true, Consumer: "function-extractor", Operation: "define-extraction-recipe", Proof: "infinite-regress"},
	}
	values := make(map[string]int, len(raw))
	seen := make(map[string]bool, len(raw))
	for _, encoded := range raw {
		var indicator extractorIndicatorRecord
		if decodeStrictBytes(encoded, &indicator) != nil || indicator.ID == "" || indicator.Value < 0 || seen[indicator.ID] {
			return nil, false
		}
		required, exists := expected[indicator.ID]
		if !exists || indicator.Limit != required.Limit || indicator.Blocking != required.Blocking ||
			indicator.Consumer != required.Consumer || indicator.Operation != required.Operation || indicator.Proof != required.Proof {
			return nil, false
		}
		seen[indicator.ID] = true
		values[indicator.ID] = indicator.Value
	}
	if len(seen) != len(expected) {
		return nil, false
	}
	for id := range expected {
		if _, ok := values[id]; !ok {
			return nil, false
		}
	}
	return values, true
}

func extractorCreatedCount(subjects []extractorSubject) int {
	count := 0
	for _, subject := range subjects {
		count += len(subject.CreatedFiles)
	}
	return count
}

func reportFailureOperationError(failures []extractorFailureRecord) (*operationError, bool) {
	var unknown *operationError
	for _, failure := range failures {
		if !validExtractorFailure(failure) {
			return newOperationError("evaluate-operation", "decode-function-extraction-report", "INSTANCE_EVIDENCE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample"), true
		}
		candidate := newOperationError(failure.Stage, failure.Step, failure.Reason, failureClass(failure), failure.NextOperation)
		candidate.blockedBy = append([]string{}, failure.BlockedBy...)
		if failure.Decision == "REFUTED" {
			return candidate, true
		}
		if unknown == nil {
			unknown = candidate
		}
	}
	if unknown != nil {
		return unknown, true
	}
	return nil, false
}

func validExtractorFailure(failure extractorFailureRecord) bool {
	if failure.Logical == "" || failure.Decision == "" || failure.Stage == "" ||
		failure.Step == "" || failure.Reason == "" || failure.NextOperation == "" ||
		failure.BlockedBy == nil {
		return false
	}
	switch failure.Decision {
	case "REFUTED":
		return failure.UnknownClass == "" && failure.BlockerID != ""
	case "UNKNOWN":
		return failure.UnknownClass == "DIRECT_MISSING" ||
			failure.UnknownClass == "MALFORMED_EVIDENCE" ||
			failure.UnknownClass == "UNEXPECTED_EVIDENCE" ||
			failure.UnknownClass == "DEPENDENCY_BLOCKED"
	default:
		return false
	}
}

func failureClass(failure extractorFailureRecord) string {
	if failure.Decision == "REFUTED" {
		return "KNOWN_CONTRADICTION"
	}
	return failure.UnknownClass
}

func validateOutputFiles(root string, subject sourcepolicy.SourceSubject, packageName string, imports []string, header []byte, observed extractorSubject) (int, int, map[string]bool, error) {
	found, afterLines := 0, 0
	seen := make(map[string]bool, len(observed.Files))
	observedImports := make(map[string]bool)
	for _, logical := range observed.Files {
		if seen[logical] {
			return 0, 0, nil, fmt.Errorf("changed output %s is duplicated", logical)
		}
		seen[logical] = true
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
		if err != nil || physicalLineCount(data) == 0 {
			return 0, 0, nil, fmt.Errorf("changed output %s is unavailable", logical)
		}
		set, file, err := parseGoFile(logical, data)
		if err != nil || file.Name.Name != packageName || !bytes.Equal(data, mustFormat(data)) {
			return 0, 0, nil, fmt.Errorf("changed output %s is not canonical", logical)
		}
		if !bytes.Equal(sourceHeader(data, set, file), header) {
			return 0, 0, nil, fmt.Errorf("changed output %s lost build header", logical)
		}
		if err := validateImports(logical, file, imports, observedImports); err != nil {
			return 0, 0, nil, err
		}
		if function, ok := functionInFile(file, logical, subject); ok {
			if declarationLinesFor(set, function) > 75 {
				return 0, 0, nil, fmt.Errorf("function %s remains oversized", subject.Name)
			}
			afterLines = declarationLinesFor(set, function)
			found++
		}
	}
	if !importSetEqual(imports, observedImports) {
		return 0, 0, nil, fmt.Errorf("changed outputs lost import identity")
	}
	return found, afterLines, seen, nil
}

func validateImports(logical string, file *ast.File, imports []string, observed map[string]bool) error {
	for _, item := range importIdentity(file) {
		if !containsString(imports, item) {
			return fmt.Errorf("changed output %s introduced import %s", logical, item)
		}
		observed[item] = true
	}
	return nil
}

func importSetEqual(expected []string, observed map[string]bool) bool {
	if len(expected) != len(observed) {
		return false
	}
	for _, item := range expected {
		if !observed[item] {
			return false
		}
	}
	return true
}

func validateBackupCleanup(observation backupCleanupObservation, replacements []namespaceReplacementReceipt) error {
	expected := 0
	for _, replacement := range replacements {
		if replacement.DestinationPreexisted {
			expected++
		}
	}
	if observation.Attempted < 0 || observation.Removed < 0 || observation.Failures < 0 ||
		observation.Attempted != expected || observation.Removed > observation.Attempted ||
		observation.Failures > observation.Attempted ||
		observation.Removed+observation.Failures > observation.Attempted {
		return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
	}
	switch observation.Status {
	case "PASS":
		if observation.Removed != observation.Attempted || observation.Failures != 0 {
			return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
		}
		return nil
	case "PENDING":
		if observation.Removed != 0 || observation.Failures != 0 {
			return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
		}
		return &extractValidationUnknown{reason: "BACKUP_CLEANUP_UNAVAILABLE"}
	case "UNKNOWN":
		if observation.Failures == 0 || observation.Removed+observation.Failures != observation.Attempted {
			return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
		}
		return &extractValidationUnknown{reason: "BACKUP_CLEANUP_UNAVAILABLE"}
	case "NOT_APPLICABLE":
		if len(replacements) != 0 || observation.Attempted != 0 || observation.Removed != 0 || observation.Failures != 0 {
			return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
		}
		return &extractValidationUnknown{reason: "BACKUP_CLEANUP_NOT_APPLICABLE"}
	default:
		return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
	}
}

func functionInFile(file *ast.File, logical string, subject sourcepolicy.SourceSubject) (*ast.FuncDecl, bool) {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name != subject.Name {
			continue
		}
		if logical == subject.Path || function.Recv == nil {
			return function, true
		}
	}
	return nil, false
}

func parseGoFile(logical string, data []byte) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, data, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, file, nil
}

func declarationLinesFor(fset *token.FileSet, node ast.Node) int {
	start := fset.Position(node.Pos()).Line
	end := fset.Position(node.End()).Line
	return end - start + 1
}

func importIdentity(file *ast.File) []string {
	result := make([]string, 0)
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok.String() != "import" {
			continue
		}
		for _, raw := range group.Specs {
			item := raw.(*ast.ImportSpec)
			name := ""
			if item.Name != nil {
				name = item.Name.Name
			}
			result = append(result, name+":"+item.Path.Value)
		}
	}
	sort.Strings(result)
	return result
}

func sourceHeader(data []byte, fset *token.FileSet, file *ast.File) []byte {
	position := fset.Position(file.Package)
	if position.Offset <= 0 || position.Offset > len(data) {
		return nil
	}
	return data[:position.Offset]
}

func mustFormat(data []byte) []byte {
	formatted, err := format.Source(data)
	if err != nil {
		return nil
	}
	return formatted
}

func containsString(values []string, value string) bool {
	return slices.Contains(values, value)
}

func extractValidationErrorReason(err error) string {
	reason := "INSTANCE_CONFORMANCE_FAILED"
	if replacementErr, ok := errors.AsType[*namespaceReplacementError](err); ok {
		reason = replacementErr.reason
	}
	if unavailable, ok := errors.AsType[*extractValidationUnknown](err); ok {
		reason = unavailable.reason
	}
	return reason
}

func extractValidationErrorClass(err error) string {
	if _, ok := errors.AsType[*extractValidationUnknown](err); ok {
		return "DIRECT_MISSING"
	}
	return "KNOWN_CONTRADICTION"
}

func extractValidationNextOperation(err error) string {
	if extractValidationErrorClass(err) == "DIRECT_MISSING" {
		return "recover-backup-cleanup-evidence"
	}
	return "report-counterexample"
}
