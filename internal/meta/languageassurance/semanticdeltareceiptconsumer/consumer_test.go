package semanticdeltareceiptconsumer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConsumerRejectsUnsealedWireReceipt(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	input := Input{CaseID: "equivalent", SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", BeforePath: filepath.Join(root, "examples/semantic-delta-receipt/before.gooo"), AfterPath: filepath.Join(root, "examples/semantic-delta-receipt/equivalent-after.gooo")}
	verdict := AdjudicateFiles(input, Receipt{})
	if verdict.Passed || verdict.Reason != reasonReceipt || verdict.Consumer != consumerName {
		t.Fatalf("verdict=%+v", verdict)
	}
}

func TestConsumerRejectsResealedLedgerTamper(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	input := Input{CaseID: "equivalent", SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", BeforePath: filepath.Join(root, "examples/semantic-delta-receipt/before.gooo"), AfterPath: filepath.Join(root, "examples/semantic-delta-receipt/equivalent-after.gooo")}
	beforeRaw, err := os.ReadFile(input.BeforePath)
	if err != nil {
		t.Fatal(err)
	}
	afterRaw, err := os.ReadFile(input.AfterPath)
	if err != nil {
		t.Fatal(err)
	}
	before, err := projectSource(input.BeforePath, beforeRaw)
	if err != nil {
		t.Fatal(err)
	}
	after, err := projectSource(input.AfterPath, afterRaw)
	if err != nil {
		t.Fatal(err)
	}
	text := textualDelta(beforeRaw, afterRaw)
	structural := structuralDelta(before, after)
	claims := claimDelta(before, after)
	class, decision, reason := classDecision(structural, claims)
	ledger, transitions := claimLedger(before, after, class)
	receipt := Receipt{Schema: receiptSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, Producer: producerName, Consumer: consumerName, MetaOperation: metaOperation, ProofChoice: "FOUNDATION", Stage: "produce", Step: "separate-delta-layers", Before: snapshot(beforeRaw, before, nil), After: snapshot(afterRaw, after, nil), TextualDelta: text, StructuralDelta: structural, SemanticClaimDelta: claims, RawDecision: text.Decision, SemanticDecision: semanticDecision(class), Decision: decision, Resolution: resolutionExact, Classification: class, Reason: reason, ClaimLedger: ledger, ClaimTransitions: transitions, RepositoryWrites: 0}
	receipt.ReceiptDigest = digestValue(receipt)
	tampered := receipt
	tampered.ClaimLedger = append([]Claim(nil), receipt.ClaimLedger...)
	tampered.ClaimLedger[0].Status = statusRefuted
	tampered.ReceiptDigest = ""
	tampered.ReceiptDigest = digestValue(tampered)
	verdict := AdjudicateFiles(input, tampered)
	if verdict.Passed || verdict.Reason != reasonTextualOnly {
		t.Fatalf("resealed tamper was accepted: %+v", verdict)
	}
}
