package main

import (
	"bytes"
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
	if err := decodeStrictBytes(raw, &report); err != nil ||
		report.Schema != functionExtractionReportSchema || report.SourceSHA != expectedSHA ||
		len(report.Unhandled) != 0 {
		return raw, extractorReport{}, fmt.Errorf("malformed extraction report")
	}
	return raw, report, nil
}

func failedExtractionError(root, reportName string, plan generation.Plan) *operationError {
	path := filepath.Join(root, reportName)
	_, report, err := decodeExtractorReport(path, plan.HeadSHA)
	if err != nil {
		if os.IsNotExist(err) {
			return &operationError{"execute-operation", "run-function-extractor", "EXECUTOR_PROCESS_FAILED", "DIRECT_MISSING", "restore-operation-evidence"}
		}
		return &operationError{"evaluate-operation", "decode-function-extraction-report", "INSTANCE_EVIDENCE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample"}
	}
	if validationErr := validateBackupCleanup(report.BackupCleanup, report.NamespaceReplacements); validationErr != nil {
		reason := extractValidationErrorReason(validationErr)
		class := extractValidationErrorClass(validationErr)
		next := extractValidationNextOperation(validationErr)
		return &operationError{"evaluate-operation", "validate-function-extraction", reason, class, next}
	}
	return &operationError{"execute-operation", "run-function-extractor", "EXECUTOR_PROCESS_FAILED", "DIRECT_MISSING", "restore-operation-evidence"}
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
	case "PENDING":
		if observation.Removed != 0 || observation.Failures != 0 {
			return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
		}
	case "UNKNOWN":
		if observation.Failures == 0 || observation.Removed+observation.Failures != observation.Attempted {
			return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
		}
	default:
		return &namespaceReplacementError{reason: "BACKUP_CLEANUP_INCONSISTENT"}
	}
	return nil
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
	var replacementErr *namespaceReplacementError
	if errors.As(err, &replacementErr) {
		reason = replacementErr.reason
	}
	var unavailable *extractValidationUnknown
	if errors.As(err, &unavailable) {
		reason = unavailable.reason
	}
	return reason
}

func extractValidationErrorClass(err error) string {
	var unavailable *extractValidationUnknown
	if errors.As(err, &unavailable) {
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
