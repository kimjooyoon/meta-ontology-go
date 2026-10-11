package bodycodegen

import (
	"strings"
	"testing"
)

const countedIntegerPrelude = `package bounded
namespace bounded
entity Integer id "gooo://std/integer"
`

func TestCountedLoopsPreserveHelperCallBudget(t *testing.T) {
	source := []byte(countedIntegerPrelude + "activity Step(Integer) -> Integer computes `return input + 1`\n" +
		"activity Run(Integer) -> Integer computes `let i = 0; let total = 0; for i = 0; i < 1024; i = i + 1 { total = Step(total) }; return total`\n")
	generated, err := Generate("budget.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	if generated.Report.CallClosure == nil || generated.Report.CallClosure.MaxCallsPerInvocation != 1024 ||
		len(generated.Report.CallClosure.Edges) != 1 {
		t.Fatal("repeated calls lost their bound or source edge identity", generated.Report.CallClosure)
	}
	results, _, err := evaluateIntegerCases([]byte(generated.Source), "Run", []IRBodyFillTestCase{{Inputs: []int64{0}, Expected: 1024}})
	if err != nil || !results[0].Passed {
		t.Fatal(results, err)
	}
	heavy := strings.Replace(string(source), "return input + 1", "return Leaf(input) + Leaf(input) + Leaf(input) + Leaf(input)", 1) +
		"activity Leaf(Integer) -> Integer computes `return input`\n"
	if _, err := Generate("over-budget.gooo", []byte(heavy), "Run"); err == nil || !strings.Contains(err.Error(), "4096 calls") {
		t.Fatal("loop's repeated helper closure escaped the existing call budget", err)
	}
}

func TestCountedLoopsRejectUnboundedOrMutableCounters(t *testing.T) {
	for _, loop := range []string{
		"for i < input { i = i + 1 }",
		"for i = 0; i < input; i = i + 1 {}",
		"for i = 0; i < 0; i = i + 1 {}",
		"for i = 0; i < 1025; i = i + 1 {}",
		"for i = 0; i < 2; i = i + 0 {}",
		"for i = 0; i <= 2; i = i + 1 {}",
		"for i = 0; i < 2; i = i + 1 { i = 0 }",
		"for i = 0; i < 2; i = i + 1 { let i = 0 }",
		"for i = 0; i < 2; i = i + 1 { let j = 0; for j = 0; j < 2; j = j + 1 {} }",
		"for input = 0; input < 2; input = input + 1 {}",
	} {
		source := []byte(countedIntegerPrelude + "activity Run(Integer) -> Integer computes `let i = 0; " + loop + "; return i`\n")
		if _, err := Generate("invalid-loop.gooo", source, "Run"); err == nil {
			t.Fatal("unsupported loop accepted", loop)
		}
	}
}

func TestCountedCallCollectionKeepsEarlierLoopFailure(t *testing.T) {
	_, err := directPureCallSites("var i = 0; for i = 0; i < input; i = i + 1 { Step(i) }; return Step(input)")
	if err == nil || !strings.Contains(err.Error(), "integer literal") {
		t.Fatal("a later valid call erased the invalid loop bound", err)
	}
}
