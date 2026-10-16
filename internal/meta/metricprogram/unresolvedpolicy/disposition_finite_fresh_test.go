package unresolvedpolicy_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestDispositionFiniteSourceFresh1753(t *testing.T) {
	base, err := os.ReadFile("policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	source.Write(base)
	source.WriteString(strings.TrimRight(dispositionRecord1751, "\n") + " assembling {\n")
	for _, row := range []struct {
		inputs, result string
	}{
		{`[true,true]`, `{"resolution":"exact_operation","reason":"REGISTERED_SOURCE_OPERATION","next_operation":"CONTINUE_SOURCE_OPERATION","readiness_credit":false,"repository_workspace_writes":false,"promotion_authorized":false}`},
		{`[false,true]`, `{"resolution":"invariant_only","reason":"OPERATION_NOT_IN_SOURCE_REGISTRY","next_operation":"DECLARE_SOURCE_OPERATION","readiness_credit":false,"repository_workspace_writes":false,"promotion_authorized":false}`},
		{`[true,false]`, `{"resolution":"invariant_only","reason":"OPERATION_ORIGIN_REQUIRED","next_operation":"CAPTURE_OPERATION_ORIGIN","readiness_credit":false,"repository_workspace_writes":false,"promotion_authorized":false}`},
		{`[false,false]`, `{"resolution":"invariant_only","reason":"OPERATION_ORIGIN_REQUIRED","next_operation":"CAPTURE_OPERATION_ORIGIN","readiness_credit":false,"repository_workspace_writes":false,"promotion_authorized":false}`},
	} {
		fmt.Fprintf(&source, "value_case %q -> %q\n", row.inputs, row.result)
	}
	source.WriteString("attempts \"1\"\n}\n")
	generator, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := generator.GenerateSourceAssembly(context.Background(), "source-disposition-finite-1753.gooo",
		[]byte(source.String()), "DescribeOperation")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.RecordAssembly
	if receipt == nil || receipt.Status != "COMPLETE_FINITE" || receipt.Total != 4 || receipt.Passed != 4 ||
		receipt.FieldsTotal != 24 || receipt.FieldsPassed != 24 || receipt.ModelRequested || receipt.ModelCalls != 0 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("finite source-owned diagnostic contract lost its bounded domain: %+v", receipt)
	}
}
