package semanticdeltareceiptconsumer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func fixture(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	return filepath.Join(root, "examples", "semantic-delta-receipt", name)
}

func TestConsumerRejectsUnsealedWireReceipt(t *testing.T) {
	input := Input{CaseID: "equivalent", SubjectSHA: candidateSHA, ObservedCheckoutSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("equivalent-after.gooo")}
	verdict := AdjudicateFiles(input, Receipt{})
	if verdict.Passed || verdict.Reason != reasonReceipt || verdict.Consumer != consumerName {
		t.Fatalf("verdict=%+v", verdict)
	}
}

func TestConsumerRejectsResealedTamperMatrix(t *testing.T) {
	cases := []struct {
		name  string
		input Input
	}{
		{name: "exact", input: Input{CaseID: "equivalent", SubjectSHA: candidateSHA, ObservedCheckoutSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("equivalent-after.gooo")}},
		{name: "parse-unknown", input: Input{CaseID: "indeterminate", SubjectSHA: candidateSHA, ObservedCheckoutSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("indeterminate-after.gooo")}},
		{name: "subject-unknown", input: Input{CaseID: "subject-unknown", SubjectSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("equivalent-after.gooo")}},
		{name: "ambiguous", input: Input{CaseID: "ambiguous-match", SubjectSHA: candidateSHA, ObservedCheckoutSHA: candidateSHA, BeforePath: fixture("ambiguous-before.gooo"), AfterPath: fixture("ambiguous-after.gooo")}},
	}
	type tamper struct {
		name string
		edit func(*Receipt)
	}
	tampers := []tamper{
		{name: "proof-choice", edit: func(r *Receipt) { r.ProofChoice = "TAMPERED" }},
		{name: "stage", edit: func(r *Receipt) { r.Stage = "TAMPERED" }},
		{name: "step", edit: func(r *Receipt) { r.Step = "TAMPERED" }},
		{name: "reason", edit: func(r *Receipt) { r.Reason = "TAMPERED" }},
		{name: "decision", edit: func(r *Receipt) { r.Decision = "TAMPERED" }},
		{name: "resolution", edit: func(r *Receipt) { r.Resolution = "TAMPERED" }},
		{name: "classification", edit: func(r *Receipt) { r.Classification = "TAMPERED" }},
		{name: "expected-subject", edit: func(r *Receipt) { r.ExpectedSubjectSHA = strings.Repeat("b", 40) }},
		{name: "observed-checkout", edit: func(r *Receipt) { r.ObservedCheckoutSHA = strings.Repeat("b", 40) }},
		{name: "meta-contract", edit: func(r *Receipt) { r.MetaContractDigest = "sha256:" + strings.Repeat("b", 64) }},
		{name: "transition-head", edit: func(r *Receipt) { r.TransitionHeadDigest = "sha256:" + strings.Repeat("b", 64) }},
		{name: "effects-status", edit: func(r *Receipt) { r.Effects.Status = "NET_REPOSITORY_STATE_CHANGED" }},
	}
	if len(tampers) != 12 {
		t.Fatalf("tamper matrix size=%d", len(tampers))
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			beforeRaw, err := os.ReadFile(testCase.input.BeforePath)
			if err != nil {
				t.Fatal(err)
			}
			afterRaw, err := os.ReadFile(testCase.input.AfterPath)
			if err != nil {
				t.Fatal(err)
			}
			expected := reconstructReceipt(testCase.input, beforeRaw, afterRaw)
			rejected := 0
			for _, mutation := range tampers {
				tampered := expected
				mutation.edit(&tampered)
				tampered.ReceiptDigest = ""
				tampered.ReceiptDigest = digestValue(tampered)
				verdict := AdjudicateFiles(testCase.input, tampered)
				if !verdict.Passed {
					rejected++
				}
			}
			if rejected != 12 {
				t.Fatalf("resealed tamper matrix rejected=%d/12", rejected)
			}
		})
	}
}
