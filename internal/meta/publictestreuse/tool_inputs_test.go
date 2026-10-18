package publictestreuse

import (
	"errors"
	"strings"
	"testing"
)

func TestReceiptBindsSelectedGoCommand(t *testing.T) {
	original := packageInputReceiptFixture(t)
	for _, field := range []string{"path", "executable", "environment", "platform"} {
		t.Run(field, func(t *testing.T) {
			changed := original.Binding
			switch field {
			case "path":
				changed.ToolInputs.Path = "/another/go"
			case "executable":
				changed.ToolInputs.ExecutableDigest = strings.Repeat("b", 64)
			case "environment":
				changed.ToolInputs.EnvironmentDigest = strings.Repeat("b", 64)
			case "platform":
				changed.ToolInputs.Platform = "another/host"
			}
			changed.ToolchainDigest = changed.ToolInputs.Digest()
			if err := VerifyReceipt(original, changed); err == nil {
				t.Fatal("changed selected tool inputs reused an old result")
			}
			forged := original
			forged.Binding.ToolInputs = changed.ToolInputs
			forged.ReceiptID, _ = ReceiptContentDigest(forged)
			if err := ValidateReceipt(forged); err == nil {
				t.Fatal("resealed contradictory toolchain identity passed")
			}
		})
	}
}

func TestHistoricalReceiptHasNoSelectedToolObservation(t *testing.T) {
	if _, err := DecodeReceipt([]byte(`{"schema":"gooo/public-test-reuse-receipt/v2"}`)); !errors.Is(err, ErrUnobservedToolInputs) {
		t.Fatal("old observer runtime must not become a selected child observation", err)
	}
	for _, data := range []string{
		`{"schema":"gooo/public-test-reuse-receipt/v2","unexpected":true}`,
		`{"schema":"gooo/public-test-reuse-receipt/v2"} {}`,
		`{"schema":"unknown","schema":"gooo/public-test-reuse-receipt/v2"}`,
	} {
		if _, err := DecodeReceipt([]byte(data)); err == nil || errors.Is(err, ErrUnobservedToolInputs) {
			t.Fatal("malformed record reclassified as an ordinary historical observation", err)
		}
	}
}
