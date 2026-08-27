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
			expected := reconstructReceipt(input, beforeRaw, afterRaw)
			rejected := 0
			for _, mutation := range mutations {
				tampered := expected
				mutation.edit(&tampered)
				tampered.ReceiptDigest = ""
				tampered.ReceiptDigest = digestValue(tampered)
				verdict := AdjudicateFiles(input, tampered)
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

func TestFixedClaimExpectationRejectsDrift(t *testing.T) {
	contract, raw, err := readClaimIdentityExpectations()
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(*claimIdentityExpectationContract)
	}{
		{name: "duplicate-case", mutate: func(value *claimIdentityExpectationContract) { value.Cases = append(value.Cases, value.Cases[0]) }},
		{name: "replaced-case", mutate: func(value *claimIdentityExpectationContract) { value.Cases[0].ID = "replacement" }},
		{name: "count-drift", mutate: func(value *claimIdentityExpectationContract) { value.Cases[0].ExpectedClaimTotal++ }},
		{name: "claim-replacement", mutate: func(value *claimIdentityExpectationContract) {
			value.Cases[0].ExpectedClaimIDs[0] = "gooo://semantic-delta/claim/object/replacement"
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			copy := contract
			copy.Cases = append([]claimIdentityExpectation(nil), contract.Cases...)
			for index := range copy.Cases {
				copy.Cases[index].ExpectedClaimIDs = append([]string(nil), contract.Cases[index].ExpectedClaimIDs...)
			}
			testCase.mutate(&copy)
			if validateClaimExpectationContract(copy) == nil {
				t.Fatal("fixed claim expectation drift was accepted")
			}
		})
	}
	unknownField := append([]byte(`{"unexpected":true,`), raw[1:]...)
	if _, err := decodeClaimIdentityExpectations(unknownField); err == nil {
		t.Fatal("unknown expectation field was accepted")
	}
}

func TestClaimIdentityUnknownCoordinates(t *testing.T) {
	base := Input{CaseID: "equivalent", SubjectSHA: candidateSHA, ObservedCheckoutSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("equivalent-after.gooo")}
	for _, testCase := range []struct {
		name       string
		input      Input
		wantStep   string
		wantReason string
	}{
		{name: "missing-before", input: Input{CaseID: base.CaseID, SubjectSHA: base.SubjectSHA, ObservedCheckoutSHA: base.ObservedCheckoutSHA, BeforePath: fixture("missing-before.gooo"), AfterPath: base.AfterPath}, wantStep: claimSourceBeforeStep, wantReason: claimSourceBeforeReason},
		{name: "missing-after", input: Input{CaseID: base.CaseID, SubjectSHA: base.SubjectSHA, ObservedCheckoutSHA: base.ObservedCheckoutSHA, BeforePath: base.BeforePath, AfterPath: fixture("missing-after.gooo")}, wantStep: claimSourceAfterStep, wantReason: claimSourceAfterReason},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			comparison := ValidateFixedClaimIdentity(testCase.input)
			if comparison.Decision != decisionFailClosed || comparison.Resolution != resolutionLower || comparison.Stage != "source-pair" || comparison.Step != testCase.wantStep || comparison.Reason != testCase.wantReason {
				t.Fatalf("comparison=%+v", comparison)
			}
		})
	}
}
