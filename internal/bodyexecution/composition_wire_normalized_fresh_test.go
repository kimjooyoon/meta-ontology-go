package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCompositionWireNormalizedBudgetFresh1881(t *testing.T) {
	source, _, value := wireFixture1874(t)
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	suite := CompositionCases{Schema: CompositionInputsSchema}
	for range 32 {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"First": raw}})
	}
	err = ValidateCompositionCases(context.Background(), "wire.gooo", source, suite)
	if err == nil || !strings.Contains(err.Error(), "normalized suite JSON") || !strings.Contains(err.Error(), "maximum 1048576 bytes") {
		t.Fatal("in-memory suite bypassed normalized transport admission", err)
	}
	if len(suite.Cases) != 32 {
		t.Fatal("budget refusal silently removed rows")
	}
}
