package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestPolicyRevisionOperationNativeCLIAndIndependentReceipt(t *testing.T) {
	source, request := revisionObservationFixture(t)
	rawRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	policyPath := writeRevisionOperationInput(t, temp, "policy.gooo", source)
	requestPath := writeRevisionOperationInput(t, temp, "request.json", rawRequest)
	contract := PolicyRevisionOperationContract()
	operationPath := writeRevisionOperationInput(t, temp, "operation.gooo", contract)
	output := runRevisionOperationNativeCLI(t, root, "./cmd/meta-policy-compilation-witness",
		"-policy", policyPath, "-observe-revision", requestPath, "-revision-operation", operationPath,
		"-profile-package", "metapolicycompilation", "-profile-namespace", "metapolicycompilation")
	var report PolicyRevisionOperationObservation
	if err := json.Unmarshal(output, &report); err != nil {
		t.Fatalf("decode Gooo-bound CLI observation: %v", err)
	}
	if report.Schema != PolicyRevisionOperationSchema || report.NativeWorkerInvocations != 1 ||
		report.Observation == nil || report.Binding.Program != PolicyRevisionOperationProgram ||
		!report.Binding.UsedPolicySource || !report.Binding.UsedRevisionRequest || !report.Binding.GeneratedObservation {
		t.Fatalf("CLI did not execute its bound operation: %+v", report)
	}
	observation := report.Observation
	if observation.ExecutionStatus != "COMPLETED" || observation.ExecutionConformance != "PASS" ||
		observation.Counts.RequestedCasePairs != len(request.Cases) ||
		observation.Counts.ObservedCasePairs != len(request.Cases) ||
		observation.Counts.ReplayComparisons != 2*len(request.Cases) ||
		observation.RequestArtifactDigest != DigestBytes(rawRequest) ||
		observation.Admission.State != "UNKNOWN" || observation.Improvement != "UNKNOWN" ||
		report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
		t.Fatalf("bound operation changed execution/accounting/adoption semantics: %+v", report)
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(output, &wire); err != nil {
		t.Fatal(err)
	}
	receiptPath := writeRevisionOperationInput(t, temp, "receipt.json", wire["observation"])
	independent := runRevisionOperationNativeCLI(t, root, "./cmd/meta-policy-compilation-consumer",
		"-policy", policyPath, "-revision-request", requestPath, "-observe-revision-receipt", receiptPath,
		"-profile-package", "metapolicycompilation", "-profile-namespace", "metapolicycompilation")
	assertRevisionOperationIndependentReceipt(t, independent, len(request.Cases))
	for path, before := range map[string][]byte{policyPath: source, requestPath: rawRequest, operationPath: contract} {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("caller input changed: path=%s error=%v", path, err)
		}
	}
	t.Logf("Gooo activity=%s; native worker invocations=%d; observed pairs=%d; replay comparisons=%d; admission=%s",
		report.Binding.ActivityID, report.NativeWorkerInvocations, observation.Counts.ObservedCasePairs,
		observation.Counts.ReplayComparisons, observation.Admission.State)
}

func writeRevisionOperationInput(t *testing.T, directory, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runRevisionOperationNativeCLI(t *testing.T, root, program string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", append([]string{"run", program}, args...)...)
	command.Dir = root
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	if err := command.Run(); err != nil {
		t.Fatalf("native CLI %s: %v\n%s", program, err, diagnostic.String())
	}
	return append([]byte(nil), output.Bytes()...)
}

func assertRevisionOperationIndependentReceipt(t *testing.T, raw []byte, pairs int) {
	t.Helper()
	var report map[string]json.RawMessage
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	var decision, improvement string
	var comparisons, mutation, promotion int
	var executionObserved bool
	fields := map[string]any{
		"decision": &decision, "improvement": &improvement,
		"independent_result_comparisons": &comparisons, "mutation_authority": &mutation,
		"promotion_authority": &promotion, "policy_execution_observed": &executionObserved,
	}
	for name, target := range fields {
		if err := json.Unmarshal(report[name], target); err != nil {
			t.Fatalf("independent receipt field %s: %v", name, err)
		}
	}
	if decision != "RECEIPT_CONSISTENT_ONLY" || comparisons != 6*pairs ||
		improvement != "UNKNOWN" || mutation != 0 || promotion != 0 || executionObserved {
		t.Fatalf("independent receipt was not preserved: %s", raw)
	}
}
