package bodycodegen

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

const callBudgetPrelude = "package budget\nnamespace budget\nentity Integer id \"urn:budget:integer\"\n" +
	"activity Leaf(Integer) -> Integer id \"urn:budget:leaf\" computes `return input`\n" +
	"activity Step(Integer) -> Integer id \"urn:budget:step\" computes `return Leaf(input) + Leaf(input) + Leaf(input) + Leaf(input)`\n"

func TestPureCallBudgetBoundaryAndSourceIdentity(t *testing.T) {
	for _, row := range []struct {
		name, before           string
		prefix, ordinal, after int
	}{
		{"exact_limit", "Leaf(input)", 1, 1, 4096},
		{"one_over", "Leaf(input) + Leaf(input)", 2, 2, 4097},
		{"cached_callee", "Step(input)", 5, 1, 4100},
		{"primitive_ordinal", "int64(0) + Leaf(input) + Leaf(input)", 2, 3, 4097},
	} {
		t.Run(row.name, func(t *testing.T) {
			body := fmt.Sprintf("let total = %s; for let i = 0; i < 819; i++ { total = Step(total) }; return total", row.before)
			source := callBudgetPrelude + "activity Run(Integer) -> Integer id \"urn:budget:run\" computes `" + body + "`\n"
			result, err := Generate("budget.gooo", []byte(source), "Run")
			if row.after == 4096 {
				if err != nil || result.Report.CallClosure == nil || result.Report.CallClosure.MaxCallsPerInvocation != 4096 {
					t.Fatal("the existing inclusive limit changed", result.Report.CallClosure, err)
				}
				return
			}
			failure, ok := errors.AsType[*PureCallBudgetError](err)
			if !ok || result.Source != "" {
				t.Fatal("missing typed failure or partial generated code escaped", err)
			}
			if failure.CallerID != "urn:budget:run" || failure.CalleeID != "urn:budget:step" ||
				failure.CallOrdinal != row.ordinal || failure.PrefixBefore != row.prefix ||
				failure.PrefixAfter != row.after || failure.Multiplicity != 819 || failure.CalleeBound != 4 ||
				failure.Limit != 4096 || failure.MultiplicityCapped {
				t.Fatal("static prefix arithmetic or source identity differs", failure)
			}
		})
	}
}

func TestPureCallBudgetLocatesNestedFailureAndStopsAtFirstExcess(t *testing.T) {
	source := callBudgetPrelude +
		"activity Inner(Integer) -> Integer id \"urn:budget:inner\" computes `let total = input; for let i = 0; i < 1024; i++ { total = Step(total) }; return Leaf(total)`\n" +
		"activity Outer(Integer) -> Integer id \"urn:budget:outer\" computes `return Inner(input)`\n"
	_, err := Generate("nested.gooo", []byte(source), "Outer")
	failure, ok := errors.AsType[*PureCallBudgetError](fmt.Errorf("outer: %w", err))
	if !ok || failure.Caller != "Inner" || failure.CallerID != "urn:budget:inner" ||
		failure.PrefixAfter != 5120 || failure.CallOrdinal != 0 || failure.CalleeID != "urn:budget:step" {
		t.Fatal("nested error lost the source owner or counted unvisited calls", err)
	}
	for _, text := range []string{"4096 calls per invocation", "urn:budget:inner", "urn:budget:step", "5120"} {
		if !strings.Contains(err.Error(), text) {
			t.Errorf("text diagnostic omits %q: %v", text, err)
		}
	}
}
