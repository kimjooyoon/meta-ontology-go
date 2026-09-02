package languagetest

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Observe(request Request) Receipt {
	receipt := Receipt{Schema: ReceiptSchema, Resolution: ResolutionExact, Filename: request.Filename, SourceDigest: digestBytes([]byte(request.Source)), Cases: []Case{}, Diagnostics: []Diagnostic{}, Effects: Effects{}, NonClaims: nonClaims()}
	if strings.TrimSpace(request.Filename) == "" || request.Source == "" {
		return reject(receipt, "REQUEST", "LANGUAGE_TEST_REQUEST_INVALID", "filename and source are required")
	}
	file, diagnostics := syntax.ParseFile(request.Filename, request.Source)
	if file == nil || diagnostics.HasErrors() {
		return reject(receipt, "PARSE", "LANGUAGE_TEST_SOURCE_INVALID", "source has syntax diagnostics")
	}
	specifications, entities, err := discover(file)
	if err != nil {
		return reject(receipt, "DISCOVER", "LANGUAGE_TEST_DECLARATION_INVALID", err.Error())
	}
	if len(specifications) == 0 {
		return reject(receipt, "DISCOVER", "LANGUAGE_TESTS_MISSING", "no Gooo language test marker is declared")
	}
	receipt.Summary.Declared = len(specifications)
	for _, specification := range specifications {
		execution := sourceexecution.Execute(sourceexecution.Request{Filename: request.Filename, Source: request.Source, Entry: specification.entry})
		testCase := Case{
			Name: specification.name, MarkerID: specification.markerID, Entry: specification.entry,
			Assertion: "OUTPUT_ENTITY", Expected: entities[specification.expected],
			ExecutionDigest: execution.Digest, Decision: DecisionFailClosed,
		}
		receipt.Summary.Executed++
		if execution.Decision != sourceexecutionDecisionPass {
			testCase.Reason = "LANGUAGE_TEST_EXECUTION_FAILED"
			receipt.Summary.Failed++
			receipt.Diagnostics = append(receipt.Diagnostics, Diagnostic{Stage: "EXECUTE", Code: testCase.Reason, Message: fmt.Sprintf("activity %q did not produce an execution receipt", specification.entry)})
			receipt.Cases = append(receipt.Cases, testCase)
			continue
		}
		if receipt.SemanticDigest == "" {
			receipt.SemanticDigest = execution.SemanticDigest
		}
		testCase.Observed = Binding{Name: execution.Entry.Output.Name, ID: execution.Entry.Output.ID}
		if testCase.Observed == testCase.Expected {
			testCase.Decision = DecisionPass
			testCase.Reason = "LANGUAGE_TEST_ASSERTION_PASSED"
			receipt.Summary.Passed++
		} else {
			testCase.Reason = "LANGUAGE_TEST_ASSERTION_FAILED"
			receipt.Summary.Failed++
			receipt.Diagnostics = append(receipt.Diagnostics, Diagnostic{Stage: "ASSERT", Code: testCase.Reason, Message: fmt.Sprintf("activity %q produced %q, expected %q", specification.entry, testCase.Observed.Name, testCase.Expected.Name)})
		}
		receipt.Cases = append(receipt.Cases, testCase)
	}
	if receipt.Summary.Failed == 0 {
		receipt.Decision = DecisionPass
		receipt.Reason = "LANGUAGE_TESTS_PASSED"
	} else {
		receipt.Decision = DecisionFailClosed
		receipt.Reason = receipt.Diagnostics[0].Code
	}
	return seal(receipt)
}

const sourceexecutionDecisionPass = "PASS"

func reject(receipt Receipt, stage, code, message string) Receipt {
	receipt.Decision = DecisionFailClosed
	receipt.Reason = code
	receipt.Diagnostics = []Diagnostic{{Stage: stage, Code: code, Message: message}}
	return seal(receipt)
}
