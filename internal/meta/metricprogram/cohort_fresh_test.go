package metricprogram

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func TestDeclaredCohortTerminalsFresh1671(t *testing.T) {
	want := []CohortTerminalSpec{
		{"terminate-at-fixed-point", "replay-counterfactual", "REGRESSION", "REGRESS"},
		{"preserve-non-promoting-terminal", "preserve-non-promoting-terminal", "REGRESSION", "REGRESS"},
	}
	if CohortTerminalCount(0) != 2 || !reflect.DeepEqual(CanonicalCohortTerminals(), want) {
		t.Fatal("source-declared terminal roles changed")
	}
	if err := VerifyCompiledRegistry(); err != nil {
		t.Fatal(err)
	}
}

func TestCohortTerminalBoundsFresh1671(t *testing.T) {
	for _, index := range []int64{-1, 2} {
		t.Run(strconv.FormatInt(index, 10), func(t *testing.T) {
			raw, err := json.Marshal(CohortTerminalAt(index))
			if err != nil {
				t.Fatal(err)
			}
			var row CohortTerminalSpec
			if err := json.Unmarshal(raw, &row); err != nil || row != (CohortTerminalSpec{}) {
				t.Fatalf("out-of-contract index acquired a terminal: %s, %v", raw, err)
			}
		})
	}
}

func TestCohortTerminalOwnershipFresh1671(t *testing.T) {
	first := CanonicalCohortTerminals()
	first[0].Carrier = "changed-by-caller"
	if CanonicalCohortTerminals()[0].Carrier != "replay-counterfactual" {
		t.Fatal("a caller changed the compiled source contract")
	}
}
