package generation

import (
	"strings"
	"testing"
)

func TestCallbackExtractionObserverScopeCannotGrantAdmission(t *testing.T) {
	contract, err := LoadCallbackExtractionContract()
	if err != nil {
		t.Fatal(err)
	}
	record, err := contract.BuildRecord(4, "UNKNOWN", "native-runtime-evidence", 2, 2)
	if err != nil || len(record.Fields) != 9 || record.Fields["Scope"] != "PACKAGE_TEST_EVENTS_ONLY" ||
		record.Fields["State"] != "UNKNOWN" || record.Fields["ObservationDecision"] != "UNKNOWN" {
		t.Fatalf("bounded observer record=%+v err=%v", record, err)
	}
	if _, err := contract.BuildRecord(4, "CLOSED", "native-runtime-evidence", 2, 2); err == nil {
		t.Fatal("finite package tests granted semantic admission")
	}
	changed := strings.Replace(string(callbackExtractionContractSource), "scope=PACKAGE_TEST_EVENTS_ONLY", "scope=SEMANTIC_EQUIVALENCE", 1)
	if _, err := parseCallbackExtractionContract([]byte(changed)); err == nil {
		t.Fatal("observer scope mutation was accepted")
	}
}
