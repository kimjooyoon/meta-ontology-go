package completeness

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

func TestDeclarationGeneratesExactStructureAndJSONSchema(t *testing.T) {
	p, err := receiptprojection.Compile("receipt.gooo", DeclarationSource(), "CompletenessReceipt")
	if err != nil {
		t.Fatal(err)
	}
	for name, generate := range map[string]func() ([]byte, error){"receipt.generated.go": p.Go, "receipt.schema.json": p.JSONSchema} {
		got, err := generate()
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("generated %s differs from the authoritative Gooo declaration", name)
		}
	}
	if !DeclarationMatchesProjection() {
		t.Fatal("embedded source binding differs")
	}
	types := map[string]reflect.Type{"CompletenessReceipt": reflect.TypeFor[CompletenessReceipt](), "CompletenessDimension": reflect.TypeFor[CompletenessDimension](), "UnresolvedCompletenessClaim": reflect.TypeFor[UnresolvedCompletenessClaim]()}
	for _, e := range p.Entities {
		typ := types[e.Name]
		if typ.NumField() != len(e.Fields) {
			t.Fatal("generated/reverse field count differs")
		}
		for i, f := range e.Fields {
			field := typ.Field(i)
			observed := strings.ReplaceAll(strings.ReplaceAll(field.Type.String(), "completeness.", ""), "interface {}", "any")
			if field.Name != f.GoName || observed != f.GoType || field.Tag.Get("json") != f.Name || field.Tag.Get("gooo") != f.ID {
				t.Fatalf("reverse-observed Go field lost declaration identity: %s (%s)", f.ID, observed)
			}
		}
	}
}

func testReceipt() CompletenessReceipt {
	c := UnresolvedCompletenessClaim{ID: "reverse_observation_coverage", Status: "UNKNOWN", Reason: "no generated runtime observation", NextOperation: "observe generated runtime and bind its source"}
	return CompletenessReceipt{
		Schema: CompletenessReceiptSchema, ProfileID: "independent-input/v1",
		Decision: "PROGRESS", DecisionBasis: "only recorded evidence",
		Scope:          map[string]any{"receipt_declaration": ContractBinding(), "large_integer": json.Number("9007199254740993")},
		CoreDimensions: []string{"generation_coverage"},
		Dimensions: []CompletenessDimension{
			{ID: "generation_coverage", Status: "PASS", Numerator: 1, Denominator: 1,
				Unit: "source-bound artifact", Reason: "observed generation", Evidence: []string{"independent-output-digest"}},
			{ID: c.ID, Status: c.Status, Denominator: 1, Unit: "runtime observations",
				Reason: c.Reason, Evidence: []string{"runtime execution unavailable"}},
		},
		StatusCounts:    map[string]int{"PASS": 1, "PROGRESS": 0, "UNKNOWN": 1, "FAIL_CLOSED": 0},
		FirstUnresolved: &c, UnresolvedClaims: []UnresolvedCompletenessClaim{c}, NotClaimed: []string{"unobserved runtime"},
	}
}

func TestIndependentConsumerPreservesUnknownFrontier(t *testing.T) {
	r := testReceipt()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(append(data, '\n'))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.FirstUnresolved.Status != "UNKNOWN" || decoded.Dimensions[1].Status != "UNKNOWN" || decoded.AggregateCompletenessScore != nil {
		t.Fatal("unknown was converted to a number or success")
	}
	again, err := json.Marshal(decoded)
	if err != nil || !bytes.Equal(data, again) {
		t.Fatal("shared receipt round trip changed evidence or a large integer")
	}
}

func TestRejectAccountingAndFrontierMutation(t *testing.T) {
	for name, mutate := range map[string]func(*CompletenessReceipt){"score": func(r *CompletenessReceipt) { v := 100.0; r.AggregateCompletenessScore = &v }, "count": func(r *CompletenessReceipt) { r.StatusCounts["PASS"]++ }, "unknown-state": func(r *CompletenessReceipt) { r.Dimensions[1].Status = "SUCCESS" }, "lost-cause": func(r *CompletenessReceipt) { r.UnresolvedClaims[0].Reason = "" }, "first-stage": func(r *CompletenessReceipt) { r.FirstUnresolved.ID = "generation_coverage" }, "missing-core": func(r *CompletenessReceipt) { r.CoreDimensions = []string{"missing"} }, "zero-proof": func(r *CompletenessReceipt) { r.Dimensions[0].Numerator = 0; r.Dimensions[0].Denominator = 0 }, "source-binding": func(r *CompletenessReceipt) { r.Scope["receipt_declaration"] = "unbound" }} {
		t.Run(name, func(t *testing.T) {
			r := testReceipt()
			mutate(&r)
			if Validate(&r) == nil {
				t.Fatal("invalid receipt was accepted")
			}
		})
	}
	r := testReceipt()
	data, _ := json.Marshal(r)
	for _, bad := range [][]byte{append(bytes.Clone(data), []byte("{}")...), []byte(strings.Replace(string(data), `"profile_id":`, `"unknown":1,"profile_id":`, 1)), []byte(strings.Replace(string(data), `"profile_id":`, `"profile_id":"overwritten","profile_id":`, 1)), bytes.Repeat([]byte("x"), 1<<20+1)} {
		if _, err := Decode(bad); err == nil {
			t.Fatal("trailing, duplicate, unknown or oversized input accepted")
		}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "aggregate_completeness_score")
	missing, _ := json.Marshal(object)
	if _, err := Decode(missing); err == nil {
		t.Fatal("missing nullable key silently converted to null")
	}
	if _, err := Decode([]byte(strings.Replace(string(data), `"not_claimed":["unobserved runtime"]`, `"not_claimed":null`, 1))); err == nil {
		t.Fatal("null required array accepted")
	}
}

func TestExportedDeclarationCopyCannotMutateRuntimeBinding(t *testing.T) {
	copy := DeclarationSource()
	copy[0] ^= 1
	if !DeclarationMatchesProjection() {
		t.Fatal("caller mutated embedded declaration")
	}
}
