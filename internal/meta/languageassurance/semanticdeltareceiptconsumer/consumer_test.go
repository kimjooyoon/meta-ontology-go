package semanticdeltareceiptconsumer

import (
	"os"
	"path/filepath"
	"runtime"
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
	contexts := FixedReplayContexts()
	observedContextIDs := make([]string, 0, len(contexts))
	for _, context := range contexts {
		observedContextIDs = append(observedContextIDs, context.ID)
	}
	if err := exactIDInventory(FixedReplayContextIDs(), observedContextIDs); err != nil {
		t.Fatalf("replay context inventory drift: %v", err)
	}
	mutations := fixedTamperMutations()
	observedTamperIDs := make([]string, 0, len(mutations))
	for _, mutation := range mutations {
		observedTamperIDs = append(observedTamperIDs, mutation.id)
	}
	if err := exactIDInventory(FixedTamperIDs(), observedTamperIDs); err != nil {
		t.Fatalf("tamper inventory drift: %v", err)
	}
	for _, context := range contexts {
		t.Run(context.ID, func(t *testing.T) {
			observedCheckoutSHA := candidateSHA
			if !context.RequiresCheckoutSHA {
				observedCheckoutSHA = ""
			}
			input := Input{CaseID: context.ID, SubjectSHA: candidateSHA, ObservedCheckoutSHA: observedCheckoutSHA, BeforePath: fixture(filepath.Base(context.BeforePath)), AfterPath: fixture(filepath.Base(context.AfterPath))}
			beforeRaw, err := os.ReadFile(input.BeforePath)
			if err != nil {
				t.Fatal(err)
			}
			afterRaw, err := os.ReadFile(input.AfterPath)
			if err != nil {
				t.Fatal(err)
			}
			expected := reconstructReceipt(testCase.input, beforeRaw, afterRaw)
			rejected := 0
			for _, mutation := range mutations {
				tampered := expected
				mutation.edit(&tampered)
				tampered.ReceiptDigest = ""
				tampered.ReceiptDigest = digestValue(tampered)
				verdict := AdjudicateFiles(testCase.input, tampered)
				if !verdict.Passed {
					rejected++
				}
			}
			if rejected != TamperFixedTotal {
				t.Fatalf("resealed tamper matrix rejected=%d/%d", rejected, TamperFixedTotal)
			}
		})
	}
}

func TestFixedInventoriesRejectDrift(t *testing.T) {
	expectedTamper := FixedTamperIDs()
	expectedContexts := FixedReplayContextIDs()
	for _, testCase := range []struct {
		name     string
		expected []string
		observed []string
	}{
		{name: "tamper-missing", expected: expectedTamper, observed: expectedTamper[:len(expectedTamper)-1]},
		{name: "tamper-duplicate", expected: expectedTamper, observed: append(append([]string(nil), expectedTamper[:len(expectedTamper)-1]...), expectedTamper[0])},
		{name: "tamper-extra", expected: expectedTamper, observed: append(append([]string(nil), expectedTamper...), "unrelated")},
		{name: "tamper-replaced", expected: expectedTamper, observed: append(append([]string(nil), expectedTamper[:len(expectedTamper)-1]...), "unrelated")},
		{name: "context-missing", expected: expectedContexts, observed: expectedContexts[:len(expectedContexts)-1]},
		{name: "context-duplicate", expected: expectedContexts, observed: append(append([]string(nil), expectedContexts[:len(expectedContexts)-1]...), expectedContexts[0])},
		{name: "context-extra", expected: expectedContexts, observed: append(append([]string(nil), expectedContexts...), "unrelated")},
		{name: "context-replaced", expected: expectedContexts, observed: append(append([]string(nil), expectedContexts[:len(expectedContexts)-1]...), "unrelated")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if exactIDInventory(testCase.expected, testCase.observed) == nil {
				t.Fatal("inventory drift was accepted")
			}
		})
	}
}
