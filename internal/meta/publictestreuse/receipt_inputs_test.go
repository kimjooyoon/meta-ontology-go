package publictestreuse

import (
	"errors"
	"strings"
	"testing"
)

func TestReceiptBindsObservedPackageInputs(t *testing.T) {
	receipt := packageInputReceiptFixture(t)
	if err := VerifyReceipt(receipt, receipt.Binding); err != nil {
		t.Fatal(err)
	}
	changed := receipt.Binding
	changed.PackageInputs = ExecutionPackageInputsFromBytes(
		[]byte("module changed\n"), []byte("package fixture\n"), []byte("package fixture\n// test\n"))
	if err := VerifyReceipt(receipt, changed); err == nil {
		t.Fatal("changed module bytes reused an old receipt")
	}
	tampered := receipt
	tampered.Binding.PackageInputs.Members[1].Bytes++
	if err := ValidateReceipt(tampered); err == nil {
		t.Fatal("package input observation was outside the receipt content identity")
	}
	for _, member := range []int{0, 2} {
		t.Run(receipt.Binding.PackageInputs.Members[member].Role, func(t *testing.T) {
			contradictory := receipt
			contradictory.Binding.PackageInputs.Members[member].Digest = strings.Repeat("b", 64)
			var err error
			contradictory.ReceiptID, err = ReceiptContentDigest(contradictory)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateReceipt(contradictory); err == nil {
				t.Fatal("resealed contradictory selected-file digests were accepted")
			}
		})
	}
}

func TestLegacyReceiptCannotAuthorizeReuse(t *testing.T) {
	legacy := []byte(`{"schema":"gooo/public-test-reuse-receipt/v1"}`)
	if _, err := DecodeReceipt(legacy); !errors.Is(err, ErrUnobservedPackageInputs) {
		t.Fatalf("legacy schema without package observations: %v", err)
	}
	for _, data := range []string{
		`{"schema":"gooo/public-test-reuse-receipt/v1","unexpected":true}`,
		string(legacy) + ` {}`,
		`{"schema":"gooo/public-test-reuse-receipt/v2"}`,
	} {
		if _, err := DecodeReceipt([]byte(data)); err == nil || errors.Is(err, ErrUnobservedPackageInputs) {
			t.Fatalf("invalid current or malformed historical record = %v", err)
		}
	}
}

func packageInputReceiptFixture(t *testing.T) Receipt {
	t.Helper()
	digest := strings.Repeat("a", 64)
	inputs := ExecutionPackageInputsFromBytes(
		[]byte("module fixture\n"), []byte("package fixture\n"), []byte("package fixture\n// test\n"))
	binding := Binding{
		PolicySourceDigest: digest, PolicySemanticDigest: digest, PolicyEvaluatorDigest: digest,
		CanonicalSourceDigest: digest, CanonicalSemanticDigest: digest,
		GeneratedOutputDigest: inputs.Members[2].Digest, GeneratedSemanticDigest: digest,
		GeneratedManifestDigest: digest, CompilerDigest: digest, ReleasedToolDigest: digest,
		ToolchainDigest: digest, ToolchainVersion: "fixture", TestCommand: "fixture test",
		TestContractDigest: inputs.Members[0].Digest, PackageInputs: inputs,
	}
	binding.TestCommandDigest = TestCommandDigest(binding.TestCommand, binding.TestContractDigest)
	receipt := Receipt{
		Schema: ReceiptSchema, Operation: OriginalOperation, Decision: DecisionClosed,
		Reason: ReceiptReason, Binding: binding,
		Original: OriginalExecution{
			InvocationID: digest, Operation: OriginalOperation, ResultDigest: digest,
			Successful: true, BuildExecutions: 1, TestExecutions: 1,
			BuildMS: 1, TestMS: 1, WallMS: 2, PeakRSSKib: 1,
		},
	}
	var err error
	receipt.ReceiptID, err = ReceiptContentDigest(receipt)
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}
