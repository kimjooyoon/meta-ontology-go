package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func registrationTestCommand(t *testing.T, root, program string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, program, args...)
	command.Dir, command.Env = root, registrationEnvironment(root)
	raw, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("native command %s %v failed: %v\n%s", program, args, err, raw)
	}
	return raw
}

func registrationTestWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}

func registrationTestReadJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		t.Fatal(err)
	}
}

func registrationTestRequest(t *testing.T, root string) syntaxregistration.Request {
	t.Helper()
	paths, err := fs.Glob(os.DirFS(root),
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v*.json")
	if err != nil {
		t.Fatal(err)
	}
	version := 0
	for _, path := range paths {
		var value struct {
			Version int `json:"version"`
		}
		registrationTestReadJSON(t, filepath.Join(root, filepath.FromSlash(path)), &value)
		version = max(version, value.Version)
	}
	request := syntaxregistration.Request{BaseVersion: version, Case: languagesyntax.CaseDefinition{
		ID: "common-native-registration-fixture", Path: "examples/common-native-registration-fixture/main.gooo",
		Kind: languagesyntax.KindValid, ExpectedDecision: languagesyntax.DecisionPass, ProofChoice: "COHERENCE",
		MetaOperation: "replay-language-syntax", Scope: languagesyntax.ScopeLanguageCapability, EntityFields: true}}
	path := filepath.Join(root, filepath.FromSlash(request.Case.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	source := "package commonregistration\nnamespace commonregistration\n\n" +
		"entity Observation id \"gooo://common-registration/observation\" fields {\n" +
		"    field State id \"gooo://common-registration/state\" type string required one\n}\n" +
		"activity Capture(Observation) -> Observation computes \"record.forward:v1\"\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return request
}

func registrationAssertNativeReceipts(t *testing.T, plan generation.Plan,
	bundle generation.OperationObservationBundle, request syntaxregistration.Request) {
	t.Helper()
	seen := make(map[sourcepolicy.Operation]bool)
	for _, receipt := range bundle.Receipts {
		if seen[receipt.Operation] || receipt.InstanceEvidence == nil {
			t.Fatal("native receipt duplicated or missing process evidence")
		}
		seen[receipt.Operation] = true
		if receipt.Operation == sourcepolicy.OperationRegisterSyntax {
			if receipt.OperationInputDigest != syntaxregistration.RequestDigest(request) || len(receipt.Indicators) != 4 {
				t.Fatalf("typed request or four obligations missing: %+v", receipt)
			}
			tampered := receipt
			tampered.OperationInputDigest = "sha256:" + plan.HeadSHA
			tampered = generation.AttachInstanceEvidence(tampered, *receipt.InstanceEvidence)
			badReceipts := append([]generation.OperationReceipt{}, bundle.Receipts...)
			for index := range badReceipts {
				if badReceipts[index].Operation == receipt.Operation {
					badReceipts[index] = tampered
				}
			}
			if report := generation.VerifyReceipts(plan, badReceipts); report.Decision == generation.ReceiptDecisionConformant {
				t.Fatal("resealed input digest substitution was accepted")
			}
		}
	}
	if len(seen) != 2 || !seen[sourcepolicy.OperationRegisterSyntax] || !seen[sourcepolicy.OperationCollapseAssign] {
		t.Fatalf("native execution did not use two real independent operations: %+v", seen)
	}
}

func registrationPublishNativeEvidence(t *testing.T, plan generation.Plan, manifest generation.ExecutionManifest,
	bundle generation.OperationObservationBundle, receipts generation.ReceiptReport, candidate syntaxregistration.Candidate,
	buildMS, executionMS int64, processOutput []byte) {
	t.Helper()
	directory := os.Getenv("GOOO_REGISTRATION_NATIVE_EVIDENCE_DIR")
	if directory == "" {
		return
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	report := map[string]any{
		"schema": "gooo/common-registration-native-evaluation/v1", "head_sha": plan.HeadSHA,
		"selected_operations": len(plan.Selected), "required_independent_operations": plan.MinimumIndependent,
		"observed_receipts": len(bundle.Receipts), "observed_failures": len(bundle.Failures),
		"decision": receipts.Decision, "registration_indicators": 4, "required_registration_indicators": 4,
		"candidate_members": candidate.Emitted, "required_candidate_members": candidate.Required,
		"artifact_roles": len(candidate.Artifacts), "required_artifact_roles": candidate.RequiredArtifacts,
		"registration_request_digest": candidate.RequestDigest, "execution_binding": candidate.ExecutionBinding,
		"replay_comparisons": bundle.ReplayComparisons, "repository_writes": 0,
		"apply_scope": "CALLER_OWNED_CI_TEMP_COPY", "semantic_admission": candidate.Admission,
		"promotion_authorized": false, "utility": "UNKNOWN", "improvement": "UNKNOWN",
		"build_wall_ms": buildMS, "execution_wall_ms": executionMS,
		"capture_note": "candidate.json is a separate read-only worker replay with the same exact request"}
	for name, value := range map[string]any{"native-evaluation.json": report, "plan.json": plan,
		"manifest.json": manifest, "observations.json": bundle, "receipts.json": receipts, "candidate.json": candidate} {
		registrationTestWriteJSON(t, filepath.Join(directory, name), value)
	}
	if err := os.WriteFile(filepath.Join(directory, "native-output.txt"), processOutput, 0600); err != nil {
		t.Fatal(err)
	}
}
