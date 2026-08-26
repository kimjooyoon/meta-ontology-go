package valueexecution

import (
	"bytes"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

type measurement struct {
	baselineReason        string
	passedCases           int
	replayedCases         int
	passedCounterexamples int
	fingerprintSensitive  bool
	coreIRFailClosed      bool
}

type evidence struct {
	sourceParsed, activityResolved, programPresent, semanticBound bool
	fingerprintSensitive, registryKnown, operandParsed            bool
	signatureSupported, valueCasesExact, outputsObserved          bool
	deterministicReplay, counterexamplesExact                     bool
	unknownFailClosed, operandFailClosed, overflowFailClosed      bool
	coreIRFailClosed                                              bool
}

func measure(filename string, source []byte, program Program, cases []CaseResult, counters []CounterexampleResult) measurement {
	baseline := bytes.Replace(source, []byte(` computes "`+program.Text+`"`), nil, 1)
	changedSource := bytes.Replace(source, []byte(program.Text), []byte("int.add:2"), 1)
	changed, changedErr := Compile(filename, changedSource, program.Activity)
	_, coreIRErr := bidir.LowerDocument(program.document)
	passed, replayed := caseCounts(cases)
	return measurement{
		baselineReason: compileReason(filename, baseline, program.Activity),
		passedCases:    passed, replayedCases: replayed, passedCounterexamples: counterexampleCount(counters),
		fingerprintSensitive: changedErr == nil && changed.SemanticFingerprint != program.SemanticFingerprint,
		coreIRFailClosed:     coreIRErr != nil && strings.Contains(coreIRErr.Error(), "semantic IR does not support declaration attributes"),
	}
}

func (measured measurement) evidence(program Program, counters []CounterexampleResult) evidence {
	return evidence{
		sourceParsed: true, activityResolved: true, programPresent: program.Text != "",
		semanticBound: program.ModelProgram == program.Text, fingerprintSensitive: measured.fingerprintSensitive,
		registryKnown: program.OperationID == "int.add" && len(operationRegistry) == 1,
		operandParsed: program.Operand == 1, signatureSupported: program.operation.Arity == 1,
		valueCasesExact: measured.passedCases == 5, outputsObserved: measured.passedCases == 5,
		deterministicReplay: measured.replayedCases == 5, counterexamplesExact: measured.passedCounterexamples == 8,
		unknownFailClosed:  counterexamplePassed(counters, "unknown-operation"),
		operandFailClosed:  counterexamplePassed(counters, "malformed-operand"),
		overflowFailClosed: counterexamplePassed(counters, "integer-overflow"), coreIRFailClosed: measured.coreIRFailClosed,
	}
}
