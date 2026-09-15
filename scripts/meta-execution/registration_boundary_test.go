package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestRegistrationVerifierCannotPromotePartialOrContradictoryOutput(t *testing.T) {
	first, second := registrationVerifierPackages[0], registrationVerifierPackages[1]
	complete := []byte("ok\t" + first + "\t1.234s\nok\t" + second + "\t0.023s\n")
	if _, err := normalizeRegistrationVerifier(complete); err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{"", "ok\t" + first + "\t0.001s\n",
		string(complete) + "FAIL\n", string(complete) + "ok\t" + first + "\t0.001s\n",
		"ok\tother/package\t0.001s\nok\t" + second + "\t0.001s\n",
		"ok\t" + first + "\t(cached)\nok\t" + second + "\t0.001s\n"} {
		if _, err := normalizeRegistrationVerifier([]byte(output)); err == nil {
			t.Fatalf("incomplete or contradictory process output accepted: %q", output)
		}
	}
}

func TestRegistrationWorkerFailurePreservesSemanticCoordinates(t *testing.T) {
	for _, class := range []string{"DIRECT_MISSING", "STALE", "AMBIGUOUS", "UNBOUNDED", "DEPENDENCY_BLOCKED"} {
		observed := syntaxregistration.Failure{State: "UNKNOWN", Stage: "INPUT", Step: "pin",
			Reason: "EXACT_INPUT_UNAVAILABLE", UnknownClass: class,
			NextOperation: "restore-input", BlockedBy: []string{"input:exact"}}
		raw, err := json.Marshal(observed)
		if err != nil {
			t.Fatal(err)
		}
		failure := registrationWorkerFailure(registrationProcess{stderr: append(raw, '\n')})
		if failure.stage != observed.Stage || failure.step != observed.Step || failure.reason != observed.Reason ||
			failure.class != class || failure.next != observed.NextOperation || len(failure.blockedBy) != 1 {
			t.Fatalf("worker failure causality lost: %+v", failure)
		}
	}
	refuted := &syntaxregistration.Failure{State: "REFUTED", Stage: "ARTIFACT", Step: "compare",
		Reason: "COUNTEREXAMPLE", NextOperation: "preserve-counterexample", BlockedBy: []string{}}
	if failure := registrationNativeFailure(refuted, "fallback"); failure.class != "KNOWN_CONTRADICTION" {
		t.Fatalf("known contradiction became UNKNOWN: %+v", failure)
	}
}

func TestRegistrationContractComparisonUsesExactNativeWireFormats(t *testing.T) {
	source, semantic := strings.Repeat("a", 64), strings.Repeat("b", 64)
	action := generation.Action{InputContractSourceDigest: source, InputContractSemanticDigest: semantic}
	candidate := syntaxregistration.Candidate{ContractDigest: "sha256:" + source, SemanticDigest: semantic}
	if !registrationContractMatches(candidate, action) {
		t.Fatal("native source digest and semantic StableHash were incorrectly compared")
	}
	for _, changed := range []syntaxregistration.Candidate{
		{ContractDigest: source, SemanticDigest: semantic},
		{ContractDigest: candidate.ContractDigest, SemanticDigest: "sha256:" + semantic},
		{ContractDigest: "sha256:" + strings.Repeat("c", 64), SemanticDigest: semantic},
		{ContractDigest: candidate.ContractDigest, SemanticDigest: strings.Repeat("c", 64)},
	} {
		if registrationContractMatches(changed, action) {
			t.Fatalf("substituted native contract accepted: %+v", changed)
		}
	}
}
