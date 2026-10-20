package publictestreuse

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func TestExecutionToolInputsRetainCanonicalIdentity(t *testing.T) {
	digest := strings.Repeat("a", 64)
	observed := toolchaincli.GeneratedTestGoInputs{Profile: ExecutionToolProfile, Path: "/sdk/bin/go",
		ExecutableDigest: digest, Version: "go1.27.3", Platform: "darwin/arm64", EnvironmentDigest: digest}
	inputs := ExecutionToolInputs(observed)
	// Independently encoded from the existing named struct, sorted JSON field
	// names and typed string values. JSON equality alone misses a type alias.
	const want = "a7d6a59e4e07e3b7e04b01ccea8686b09d188a100c863860bb59861c4d26ad87"
	if inputs.Digest() != want || inputs.Validate() != nil {
		t.Fatal("shared observation changed the domain's canonical identity", inputs.Digest())
	}
	raw, err := json.Marshal(observed)
	if err != nil {
		t.Fatal(err)
	}
	var restored ExecutionToolInputs
	if err := json.Unmarshal(raw, &restored); err != nil || restored != inputs || restored.Digest() != want {
		t.Fatal("shared observation cannot be read as the existing domain record", err)
	}
}

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
