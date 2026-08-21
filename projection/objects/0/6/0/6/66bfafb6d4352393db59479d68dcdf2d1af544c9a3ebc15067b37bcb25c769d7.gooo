package cache

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestIncrementalMeasurementReceiptIsMachineReadableAndDeterministic(t *testing.T) {
	receipt, binding := validIncrementalMeasurementReceipt()
	if err := receipt.Validate(binding); err != nil {
		t.Fatal(err)
	}
	firstDigest, err := receipt.StableDigest()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var decoded IncrementalMeasurementReceipt
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, receipt) {
		t.Fatalf("JSON round-trip changed receipt: %#v != %#v", decoded, receipt)
	}
	secondDigest, err := decoded.StableDigest()
	if err != nil || secondDigest != firstDigest {
		t.Fatalf("JSON round-trip digest = %s, %v; want %s", secondDigest, err, firstDigest)
	}
	if err := decoded.Validate(binding); err != nil {
		t.Fatal(err)
	}
}
func TestIncrementalMeasurementReceiptFailsClosedOnIdentity(t *testing.T) {
	receipt, binding := validIncrementalMeasurementReceipt()
	for name, mutate := range map[string]func(*IncrementalMeasurementReceipt, *IncrementalMeasurementBinding){
		"missing receipt baseline": func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) { r.BaselineDigest = "" },
		"mismatched receipt baseline": func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) {
			r.BaselineDigest = HashBytes([]byte("other-baseline"))
		},
		"missing expected baseline": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) { b.BaselineDigest = "" },
		"mismatched expected baseline": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) {
			b.BaselineDigest = HashBytes([]byte("other-baseline"))
		},
		"missing options": func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) { r.OptionsDigest = "" },
		"mismatched options": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) {
			b.OptionsDigest = HashBytes([]byte("other-options"))
		},
		"missing Go version":   func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) { r.GoVersion = "" },
		"mismatched toolchain": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) { b.Toolchain = "go-other" },
	} {
		t.Run(name, func(t *testing.T) {
			mutatedReceipt, mutatedBinding := cloneIncrementalReceipt(receipt, binding)
			mutate(&mutatedReceipt, &mutatedBinding)
			if err := mutatedReceipt.Validate(mutatedBinding); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("identity mutation = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}
