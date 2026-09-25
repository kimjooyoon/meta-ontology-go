package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPolicyRevisionOperationBindsExactGoooRelations(t *testing.T) {
	source := PolicyRevisionOperationContract()
	binding, err := BindPolicyRevisionOperation(source)
	if err != nil {
		t.Fatal(err)
	}
	if binding.Program != PolicyRevisionOperationProgram || binding.ActivityID == "" ||
		!binding.UsedPolicySource || !binding.UsedRevisionRequest || !binding.GeneratedObservation ||
		binding.ContractSourceDigest != DigestBytes(source) ||
		binding.NativeContractSourceDigest != DigestBytes(source) || binding.ContractSemanticDigest == "" {
		t.Fatalf("incomplete native relation binding: %+v", binding)
	}
	commented, err := BindPolicyRevisionOperation(append(append([]byte(nil), source...), []byte("\n// caller-owned spelling\n")...))
	if err != nil {
		t.Fatal(err)
	}
	if commented.ContractSemanticDigest != binding.ContractSemanticDigest ||
		commented.ContractSourceDigest == binding.ContractSourceDigest ||
		commented.NativeContractSourceDigest != binding.NativeContractSourceDigest {
		t.Fatal("source identity was confused with semantic/native identity")
	}
	source[0] = 'X'
	if bytes.Equal(source, PolicyRevisionOperationContract()) {
		t.Fatal("exported contract aliases native contract storage")
	}
}

func TestPolicyRevisionOperationRejectsUnboundDeclarations(t *testing.T) {
	source := string(PolicyRevisionOperationContract())
	cases := map[string]string{
		"empty":                 "",
		"missing-policy-input":  strings.Replace(source, "(PolicySource, RevisionRequest)", "(RevisionRequest)", 1),
		"wrong-output":          strings.Replace(source, "-> RevisionObservation", "-> PolicySource", 1),
		"unknown-program":       strings.Replace(source, PolicyRevisionOperationProgram, "repository.write:v1", 1),
		"renamed-activity":      strings.Replace(source, "activity ObservePolicyDecisionRevision", "activity OtherRevision", 1),
		"wrong-source-identity": strings.Replace(source, "gooo://policy-revision-operation/policy-source", "gooo://policy-revision-operation/unbound", 1),
		"wrong-namespace":       strings.Replace(source, "namespace policyrevisionoperation", "namespace otheroperation", 1),
		"extra-activity":        source + "\nactivity UnboundRevision(PolicySource) -> RevisionObservation\n",
		"extra-entity":          source + "\nentity ExtraInput id \"gooo://policy-revision-operation/extra\"\n",
	}
	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := BindPolicyRevisionOperation([]byte(candidate)); err == nil {
				t.Fatal("unbound operation declaration was admitted")
			}
		})
	}
}

func TestPolicyRevisionOperationRejectsRequestsBeforeNativeInvocation(t *testing.T) {
	source, request := revisionObservationFixture(t)
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		operation []byte
		source    []byte
		request   []byte
	}{
		{"unbound", []byte("not a contract"), source, raw},
		{"stale-source", PolicyRevisionOperationContract(), append(append([]byte(nil), source...), '\n'), raw},
		{"null-request", PolicyRevisionOperationContract(), source, []byte("null")},
		{"unknown-field", PolicyRevisionOperationContract(), source, []byte("{\"unexpected\":true}")},
	}
	for _, candidate := range cases {
		t.Run(candidate.name, func(t *testing.T) {
			report, err := ObserveGoooPolicyDecisionRevision(context.Background(), candidate.operation,
				"policy.gooo", candidate.source, "metapolicycompilation", "metapolicycompilation", candidate.request)
			if err == nil || report.Schema != "" || report.NativeWorkerInvocations != 0 || report.Observation != nil {
				t.Fatalf("invalid input reached native worker: report=%+v err=%v", report, err)
			}
		})
	}
}

func TestPolicyRevisionOperationPreservesFailedAttemptCausality(t *testing.T) {
	source, request := revisionObservationFixture(t)
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	report, err := ObserveGoooPolicyDecisionRevision(ctx, PolicyRevisionOperationContract(),
		"policy.gooo", source, "metapolicycompilation", "metapolicycompilation", raw)
	if err == nil || report.Schema != PolicyRevisionOperationSchema ||
		report.NativeWorkerInvocations != 1 || report.Observation == nil {
		t.Fatalf("cancelled native attempt was erased: report=%+v err=%v", report, err)
	}
	observation := report.Observation
	if observation.ExecutionStatus != "FAILED" || observation.Counts.FailedBatches != 1 ||
		observation.Admission.State != "UNKNOWN" || report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
		t.Fatalf("failed attempt became accepted execution: %+v", report)
	}
	found := false
	for _, pending := range observation.Pending {
		if pending.Stage == "BASELINE_EXECUTION" {
			found = pending.State == "UNKNOWN" && pending.Step != "" && pending.Reason != "" &&
				pending.UnknownClass != "" && pending.NextOperation != "" && pending.BlockedBy != nil
		}
	}
	if !found {
		t.Fatal("failed native operation lost its six-field causal record")
	}
}
