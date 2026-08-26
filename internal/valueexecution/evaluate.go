package valueexecution

import (
	"bytes"
	"fmt"
	"io/fs"
	"math"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

type evidence struct {
	sourceParsed          bool
	activityResolved      bool
	programPresent        bool
	semanticBound         bool
	fingerprintSensitive  bool
	registryKnown         bool
	operandParsed         bool
	signatureSupported    bool
	valueCasesExact       bool
	outputsObserved       bool
	deterministicReplay   bool
	counterexamplesExact  bool
	unknownFailClosed     bool
	operandFailClosed     bool
	overflowFailClosed    bool
	coreIRFailClosed      bool
}

func Evaluate(filesystem fs.FS, sourcePath, activity, headSHA string) Report {
	report := Report{
		Schema: ReportSchema, Decision: DecisionFailClosed, Reason: ReasonSourceReadFailed,
		Resolution: ResolutionSyntaxOnly, HeadSHA: headSHA, SourcePath: sourcePath,
		NonClaims: []string{
			"general expression language", "arbitrary value types", "core semantic IR value-program preservation",
			"runtime memory or performance bounds", "repository mutation, promotion, or automatic adoption",
		},
	}
	source, err := fs.ReadFile(filesystem, sourcePath)
	if err != nil {
		report.Reason = ReasonSourceReadFailed
		return finalize(report)
	}
	report.SourceBytes = len(source)
	report.SourceLines = countLines(source)
	report.SourceDigest = digestBytes(source)
	program, err := Compile(sourcePath, source, activity)
	if err != nil {
		report.Reason = Reason(err)
		return finalize(report)
	}
	report.Resolution = ResolutionBidirValue
	report.Activity = activity
	report.ValueProgram = program.Text
	report.ValueProgramDigest = digestBytes([]byte(program.Text))
	report.SemanticFingerprint = program.SemanticFingerprint
	report.Registry = RegistrySummary{
		RegisteredOperations: len(operationRegistry), InvokedOperations: 1, OperationIDs: operationIDs(),
	}
	report.Cases = executeCases(program)
	report.Counterexamples = executeCounterexamples(sourcePath)
	baselineReason := compileReason(sourcePath, bytes.Replace(source, []byte(` computes "`+program.Text+`"`), nil, 1), activity)
	changed, changedErr := Compile(sourcePath, bytes.Replace(source, []byte(program.Text), []byte("int.add:2"), 1), activity)
	_, coreIRErr := bidir.LowerDocument(program.document)
	coreIRFailClosed := coreIRErr != nil && strings.Contains(coreIRErr.Error(), "semantic IR does not support declaration attributes")

	passedCases, replayedCases := caseCounts(report.Cases)
	passedCounterexamples := counterexampleCount(report.Counterexamples)
	report.Improvement = Improvement{
		ID: "value-level-computation", Before: coordinate(boolInt(baselineReason == ""), 1), After: coordinate(boolInt(passedCases == len(report.Cases)), 1),
		BeforeEvidence: baselineReason, AfterEvidence: digestValue(report.Cases),
	}
	report.Summary = Summary{
		ValueCasesPassed: passedCases, ValueCasesTotal: len(report.Cases),
		CounterexamplesPassed: passedCounterexamples, CounterexamplesTotal: len(report.Counterexamples),
		ValueOutputsObserved: passedCases, DeterministicReplays: replayedCases,
		RepositoryWrites: 0, CoreIRProgramPreserved: coordinate(0, 1),
		CoreIRFailClosed: coordinate(boolInt(coreIRFailClosed), 1),
	}
	checks := evidence{
		sourceParsed: true, activityResolved: true, programPresent: program.Text != "",
		semanticBound: program.ModelProgram == program.Text,
		fingerprintSensitive: changedErr == nil && changed.SemanticFingerprint != program.SemanticFingerprint,
		registryKnown: program.OperationID == "int.add" && len(operationRegistry) == 1,
		operandParsed: program.Operand == 1, signatureSupported: program.operation.Arity == 1,
		valueCasesExact: passedCases == 5, outputsObserved: passedCases == 5,
		deterministicReplay: replayedCases == 5, counterexamplesExact: passedCounterexamples == 8,
		unknownFailClosed: counterexamplePassed(report.Counterexamples, "unknown-operation"),
		operandFailClosed: counterexamplePassed(report.Counterexamples, "malformed-operand"),
		overflowFailClosed: counterexamplePassed(report.Counterexamples, "integer-overflow"),
		coreIRFailClosed: coreIRFailClosed,
	}
	report.Indicators = buildIndicators(checks)
	report.Views = buildViews(report.Indicators)
	report.Proofs = buildProofs(report, checks)
	if allIndicatorsSatisfied(report.Indicators) && baselineReason == ReasonProgramMissing {
		report.Decision = DecisionProven
		report.Reason = ReasonExactWitness
	}
	return finalize(report)
}

func executeCases(program Program) []CaseResult {
	fixtures := []struct {
		id       string
		input    int64
		expected int64
	}{
		{"negative", -2, -1}, {"negative-to-zero", -1, 0}, {"zero", 0, 1},
		{"positive", 41, 42}, {"maximum-boundary", math.MaxInt64 - 1, math.MaxInt64},
	}
	results := make([]CaseResult, 0, len(fixtures))
	for _, fixture := range fixtures {
		actual, firstErr := program.Execute([]int64{fixture.input})
		replay, secondErr := program.Execute([]int64{fixture.input})
		results = append(results, CaseResult{
			ID: fixture.id, Input: fixture.input, Expected: fixture.expected, Actual: actual, Replay: replay,
			Passed: firstErr == nil && actual == fixture.expected,
			ReplayMatched: firstErr == nil && secondErr == nil && actual == replay,
		})
	}
	return results
}

type counterexample struct {
	id       string
	expected string
	run      func() error
}

func executeCounterexamples(filename string) []CounterexampleResult {
	validSource := valueFixture(`activity Increment(Integer) -> Integer computes "int.add:1"`)
	program, _ := Compile(filename, validSource, "Increment")
	fixtures := []counterexample{
		{"missing-program", ReasonProgramMissing, func() error { _, err := Compile(filename, valueFixture(`activity Increment(Integer) -> Integer`), "Increment"); return err }},
		{"unknown-operation", ReasonProgramUnknown, func() error { _, err := Compile(filename, valueFixture(`activity Increment(Integer) -> Integer computes "int.subtract:1"`), "Increment"); return err }},
		{"malformed-operand", ReasonOperandInvalid, func() error { _, err := Compile(filename, valueFixture(`activity Increment(Integer) -> Integer computes "int.add:not-an-int"`), "Increment"); return err }},
		{"unsupported-signature", ReasonSignatureArityUnsupported, func() error { _, err := Compile(filename, valueFixture(`activity Increment(Integer, Integer) -> Integer computes "int.add:1"`), "Increment"); return err }},
		{"unresolved-reference", ReasonSemanticBindingFailed, func() error { _, err := Compile(filename, valueFixture(`activity Increment(Missing) -> Integer computes "int.add:1"`), "Increment"); return err }},
		{"syntax-error", ReasonSourceParseFailed, func() error { _, err := Compile(filename, valueFixture(`activity Increment(Integer) ->`), "Increment"); return err }},
		{"input-arity", ReasonInputArityMismatch, func() error { _, err := program.Execute(nil); return err }},
		{"integer-overflow", ReasonIntegerOverflow, func() error { _, err := program.Execute([]int64{math.MaxInt64}); return err }},
	}
	results := make([]CounterexampleResult, 0, len(fixtures))
	for _, fixture := range fixtures {
		actual := Reason(fixture.run())
		replay := Reason(fixture.run())
		results = append(results, CounterexampleResult{
			ID: fixture.id, ExpectedReason: fixture.expected, ActualReason: actual, ReplayReason: replay,
			Passed: actual == fixture.expected, ReplayMatched: actual == replay,
		})
	}
	return results
}

func valueFixture(activity string) []byte {
	return []byte("package valuewitness\nnamespace valuewitness\n\n" +
		"entity Integer id \"gooo://value-witness/entity/integer\"\n\n" + activity + "\n")
}

func compileReason(filename string, source []byte, activity string) string {
	_, err := Compile(filename, source, activity)
	return Reason(err)
}

func countLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	return bytes.Count(source, []byte{'\n'}) + boolInt(source[len(source)-1] != '\n')
}

func caseCounts(cases []CaseResult) (int, int) {
	passed, replayed := 0, 0
	for _, result := range cases {
		if result.Passed {
			passed++
		}
		if result.ReplayMatched {
			replayed++
		}
	}
	return passed, replayed
}

func counterexampleCount(results []CounterexampleResult) int {
	passed := 0
	for _, result := range results {
		if result.Passed && result.ReplayMatched {
			passed++
		}
	}
	return passed
}

func counterexamplePassed(results []CounterexampleResult, id string) bool {
	for _, result := range results {
		if result.ID == id {
			return result.Passed && result.ReplayMatched
		}
	}
	return false
}

func coordinate(satisfied, total int) Coordinate {
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10_000 / total
	}
	return Coordinate{Satisfied: satisfied, Total: total, BasisPoints: basisPoints}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func finalize(report Report) Report {
	report.Digest = reportDigest(report)
	return report
}

func requireExactCount(label string, got, want int) error {
	if got != want {
		return fmt.Errorf("%s=%d want=%d", label, got, want)
	}
	return nil
}
