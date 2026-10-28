package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func nominalRecordFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/record-collections/assembly.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestNominalRecordAssemblySelectsOrderedItemSum(t *testing.T) {
	source := nominalRecordFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "nested.gooo", source, "Total")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.Passed != 4 || r.Total != 4 || r.SelectedMask != 0 || r.ModelCalls != 0 {
		t.Fatalf("ordered item sum: %+v", r)
	}
	for i, want := range []string{"0", "14", "28", "9007199254740993"} {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(r.Cases[i].Actual, &object); err != nil || string(object["total"]) != want {
			t.Fatalf("case %d actual=%s want=%s err=%v", i, r.Cases[i].Actual, want, err)
		}
	}
	if _, err := RealizeSourceAssembly(context.Background(), "nested.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}

func TestNominalRecordAssemblyRejectsMalformedChildBeforeSelection(t *testing.T) {
	base := string(nominalRecordFixture(t))
	old := `[{\"items\":[{\"quantity\":2,\"price\":7}]}]`
	for _, bad := range []string{
		`[{\"items\":[{\"quantity\":true,\"price\":7}]}]`,
		`[{\"items\":[{\"quantity\":2}]}]`,
		`[{\"items\":[{\"quantity\":2,\"price\":7,\"extra\":0}]}]`,
		`[{\"items\":[{\"quantity\":2,\"quantity\":3,\"price\":7}]}]`,
		`[{\"items\":[null]}]`, `[{\"items\":null}]`,
	} {
		source := strings.Replace(base, old, bad, 1)
		if source == base {
			t.Fatal("fixture replacement did not occur")
		}
		if err := ValidateSourceAssembly(context.Background(), "bad-nested.gooo", []byte(source), "Total"); err == nil {
			t.Fatalf("accepted malformed child: %s", bad)
		}
	}
}

func TestNominalRecordAssemblyReceiptDescribesActualScope(t *testing.T) {
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "nested.gooo", nominalRecordFixture(t), "Total")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.CompletenessReceipt
	domain, ok := receipt.Scope["domain_scope"].(string)
	if !ok || !strings.Contains(domain, "nominal record fields") {
		t.Fatal("nominal assembly retained scalar-only scope", domain)
	}
	limits, ok := receipt.Scope["nominal_record_limits"].(map[string]any)
	if !ok || limits["reference_depth"] != 8 || limits["logical_value_slots"] != 8192 || limits["collection_elements"] != 256 {
		t.Fatal("nominal limits are absent or differ from the runtime contract", limits)
	}
	for _, claim := range []string{"full dynamic nested-field dependency provenance", "direct nested record field assignment", "identity comparison between two present optional pointers"} {
		if !slices.Contains(receipt.NotClaimed, claim) {
			t.Fatal("unsupported nominal dimension disappeared", claim)
		}
	}
	if _, err := RealizeSourceAssembly(context.Background(), "nested.gooo", nominalRecordFixture(t), result); err != nil {
		t.Fatal(err)
	}
}
