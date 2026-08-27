package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci"
)

func TestProcessObservationRejectsShellContradictions(t *testing.T) {
	value := causalci.ConsumerAdjudication{ExitCode: 0, Result: causalci.ExecutionPass, ConsumerIdentity: "gooo://consumer/causal-ci-selection/process"}
	if processObservationConformant(value, 1, nil) {
		t.Fatal("self-reported PASS with a nonzero OS exit was accepted")
	}
	if processObservationConformant(value, 0, []byte("process failure")) {
		t.Fatal("self-reported exit 0 with failure stderr was accepted")
	}
	if !processObservationConformant(value, 0, nil) {
		t.Fatal("matching PASS process observation was rejected")
	}
}
