package generation

import (
	"strings"
	"testing"
)

func TestCallbackExtractionContractOwnsSixProposalSteps(t *testing.T) {
	contract, err := LoadCallbackExtractionContract()
	if err != nil {
		t.Fatal(err)
	}
	if len(contract.Steps) != 6 || contract.SourceDigest == "" || contract.SemanticDigest == "" {
		t.Fatalf("contract=%+v", contract)
	}
	for index := range contract.Steps {
		state, evidence, observed, required := "CLOSED", "observed-proof", 1, 1
		if index >= 4 {
			state, evidence, observed = "UNKNOWN", "", 0
		}
		record, err := contract.BuildRecord(index, state, evidence, observed, required)
		if err != nil || record.Fields["State"] != state || record.Program != contract.Steps[index].Program {
			t.Fatalf("step=%d record=%+v err=%v", index, record, err)
		}
	}
	for _, state := range []string{"CLOSED", "FIXED_POINT", "PASS"} {
		if _, err := contract.BuildRecord(5, state, "forged", 5, 5); err == nil {
			t.Fatalf("proposal accepted admission state %s", state)
		}
	}
}

func TestCallbackExtractionContractRejectsSemanticMutants(t *testing.T) {
	source := string(callbackExtractionContractSource)
	for name, changed := range map[string]string{
		"authority": strings.Replace(source, "authority=PROPOSAL_ONLY", "authority=APPLY", 1),
		"counter-field": strings.Replace(source, "ObservedCount", "InventedCount", 1),
		"binding": strings.Replace(source, "bind BindCallbackExtractionSource.result -> ProveCallbackExtractionStructure.input", "", 1),
		"extra-result": source + "\nentity OperationResult\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseCallbackExtractionContract([]byte(changed)); err == nil {
				t.Fatal("semantic contract mutation was accepted")
			}
		})
	}
}
