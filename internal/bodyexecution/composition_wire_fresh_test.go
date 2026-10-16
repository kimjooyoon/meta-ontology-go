package bodyexecution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func wireFixture1874(t *testing.T) ([]byte, []byte, map[string]any) {
	t.Helper()
	source, err := os.ReadFile("testdata/wire-capacity.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	value := map[string]any{"revision": int64(-9223372036854775808)}
	for _, key := range []string{"a", "b", "c", "d", "e", "f"} {
		value[key] = strings.Repeat("\x00", 1024)
	}
	raw, err := json.Marshal(map[string]any{"schema": CompositionInputsSchema,
		"inputs": []any{map[string]any{"First": value}}})
	if err != nil {
		t.Fatal(err)
	}
	return source, raw, value
}

func TestCompositionWireInputFresh1874(t *testing.T) {
	source, raw, value := wireFixture1874(t)
	for _, mode := range []string{"inputs", "cases", "exact_budget", "over_budget", "decoded_text_bound", "duplicate"} {
		t.Run(mode, func(t *testing.T) {
			candidate := raw
			decode := DecodeCompositionInputs
			switch mode {
			case "cases":
				candidate, _ = json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": []any{
					map[string]any{"inputs": map[string]any{"First": value}, "expected": map[string]any{"Second": value}}}})
				decode = DecodeCompositionCases
			case "exact_budget", "over_budget":
				size := CompositionInputMaxBytes
				if mode == "over_budget" {
					size++
				}
				candidate = append(append([]byte(nil), raw...), bytes.Repeat([]byte(" "), size-len(raw))...)
			case "decoded_text_bound":
				bad, _ := json.Marshal(strings.Repeat("\x00", 1025))
				if _, err := decodeTextScalar(bad); err == nil {
					t.Fatal("oversized decoded Text accepted")
				}
				return
			case "duplicate":
				candidate = bytes.Replace(raw, []byte(`"inputs":`), []byte(`"inputs":[],"inputs":`), 1)
			}
			suite, err := decode(candidate)
			if mode == "over_budget" || mode == "duplicate" {
				if err == nil {
					t.Fatal("invalid transport accepted")
				}
				if mode == "over_budget" && !strings.Contains(err.Error(), "1048577 bytes; maximum 1048576 bytes") {
					t.Fatal(err)
				}
				return
			}
			if err != nil || len(candidate) <= 32<<10 {
				t.Fatal("large input boundary", len(candidate), err)
			}
			if err := ValidateCompositionCases(context.Background(), "wire.gooo", source, suite); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCompositionWireInspectionFresh1875(t *testing.T) {
	source, _, _ := wireFixture1874(t)
	plan, err := InspectComposition(context.Background(), "wire.gooo", source, "Second")
	if err != nil {
		t.Fatal(err)
	}
	c := CompositionTransportCapacity()
	if c.Schema != "gooo/composition-transport-capacity/v1" || c.InputJSONBytes != 1048576 ||
		c.NormalizedSuiteJSONBytes != 1048576 || c.NativeStdoutBytes != 1048576 || c.SeriesJSONBytes != 16777216 ||
		c.TextUTF8Bytes != 1024 || c.Rows != 128 || plan.NativeExecutions != 0 || plan.ModelCalls != 0 {
		t.Fatal(c)
	}
}

func TestCompositionWireNativeFresh1876(t *testing.T) {
	source, raw, value := wireFixture1874(t)
	suite, err := DecodeCompositionInputs(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "wire.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "wire.gooo", source, prior, suite, nativeTool())
	if err != nil || !r.RuntimeReplayed || len(r.Runs) != 2 || len(r.Traces) != 1 || len(r.Traces[0].Deliveries) != 2 {
		encoded, _ := json.Marshal(r)
		t.Fatalf("large native values: %v %s", err, encoded)
	}
	expected, _ := json.Marshal([][]any{{value, value}})
	expected = append(expected, '\n')
	if len(expected) <= 64<<10 {
		t.Fatal("fixture did not cross old stdout boundary")
	}
	for _, run := range r.Runs {
		if !run.Completed || run.OutputTruncated || run.StdoutSHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(expected)) {
			t.Fatal(run)
		}
	}
	for _, delivery := range r.Traces[0].Deliveries {
		if len(delivery.ActualFields) != 7 {
			t.Fatal("lost field identities")
		}
		for _, field := range delivery.ActualFields {
			want, _ := json.Marshal(value[field.Name])
			if field.ID != "wire://payload/"+field.Name || !bytes.Equal(field.Value, want) {
				t.Fatal("changed exact typed value", field.Name)
			}
		}
	}
	if r.ModelCalls != 0 || r.FiniteTotal != 0 {
		t.Fatal("input-only run gained accuracy or inference")
	}
	t.Logf("fresh one row: input_bytes=%d native_stdout_bytes=%d actual_fields=14 native_runs=2 models=0 finite_total=0", len(raw), len(expected))
}
