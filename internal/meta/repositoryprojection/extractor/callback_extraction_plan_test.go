package extractor

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalPaginationFailureCarriesGoooBoundRenderedProposalCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("normal callback extraction conformance is CI-only")
	}
	root := callbackPreviewRepositoryRoot(t)
	before, err := os.ReadFile(filepath.Join(root, callbackPreviewLogicalPath))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, callbackPreviewLogicalPath)
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" || len(result.Generated) != 0 {
		t.Fatalf("normal admission boundary changed: generated=%d err=%v", len(result.Generated), err)
	}
	var proposal CallbackExtractionProposal
	found := false
	for _, diagnostic := range failure.Diagnostics {
		if raw, ok := strings.CutPrefix(diagnostic, "callback_extraction_proposal="); ok {
			if err := json.Unmarshal([]byte(raw), &proposal); err != nil {
				t.Fatal(err)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("normal failure lacks a rendered alternative: %v", failure.Diagnostics)
	}
	if proposal.SourceDigest != callbackPreviewDigest(before) || proposal.Subject != "func:"+callbackPreviewTarget ||
		proposal.Closed != 4 || proposal.Unknown != 2 || proposal.Refuted != 0 || len(proposal.Claims) != 6 ||
		proposal.OperationAdmission != "UNKNOWN" || proposal.ApplyPermission != "FORBIDDEN" {
		t.Fatalf("proposal is not source-bound or escaped admission: %+v", proposal)
	}
	if len(proposal.Artifacts) < 2 || proposal.MaximumLines > 75 || proposal.LineLimit != 75 || proposal.PartitionDigest == "" {
		t.Fatalf("normal proposal capacity=%d files=%d", proposal.MaximumLines, len(proposal.Artifacts))
	}
	generated := map[string][]byte{}
	for _, artifact := range proposal.Artifacts {
		raw := []byte(artifact.Source)
		if artifact.Lines != physicalLines(raw) || artifact.Digest != proofDigest(raw) || artifact.Lines > 75 {
			t.Fatalf("artifact observation differs: %s", artifact.Path)
		}
		generated[artifact.Path] = raw
	}
	if proposal.PackageDigest != proofDigest(generatedPackagePayload(generated)) {
		t.Fatal("final package digest differs")
	}
	if err := projectedFinalConformance(root, callbackPreviewLogicalPath, generated); err != nil {
		t.Fatal(err)
	}
	direct, dependent := proposal.Claims[4], proposal.Claims[5]
	if direct.UnknownClass != "DIRECT_MISSING" || len(direct.BlockedBy) != 0 || direct.Stage == "" || direct.Step == "" ||
		direct.Reason == "" || direct.NextOperation == "" || dependent.UnknownClass != "DEPENDENCY_BLOCKED" ||
		len(dependent.BlockedBy) != 1 || dependent.BlockedBy[0] != direct.ID {
		t.Fatalf("unknown causal frontier differs: direct=%+v dependent=%+v", direct, dependent)
	}
	t.Logf("normal proposal claims=%d/%d/%d files=%d max_lines=%d limit=%d meta_activities=%d admission=%s",
		proposal.Closed, proposal.Unknown, proposal.Refuted, len(proposal.Artifacts), proposal.MaximumLines,
		proposal.LineLimit, len(proposal.Contract.Steps), proposal.OperationAdmission)
}

func TestCallbackExtractionRejectsUnsupportedCaptureCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("callback extraction capture conformance is CI-only")
	}
	source := "package fixture\nimport h \"net/http\"\nimport \"testing\"\nfunc " + callbackPreviewTarget + "(t *testing.T) {\n" + `
const marker = "capture"
callback := h.HandlerFunc(func(_ h.ResponseWriter, _ *h.Request) {
if marker == "" { panic("missing marker") }
})
callback(nil, nil)
}` + "\n"
	root := writeCallbackFactoryFixture(t, source)
	proposal, err := PlanCallbackExtraction(root, "fixture_test.go", "func:"+callbackPreviewTarget)
	if err == nil || !strings.Contains(err.Error(), "unsupported non-variable callback capture marker") || len(proposal.Artifacts) != 0 {
		t.Fatalf("unsupported capture escaped planning: artifacts=%d err=%v", len(proposal.Artifacts), err)
	}
}

func TestCallbackExtractionFinalPartitionRejectsMutation(t *testing.T) {
	intermediate := []byte("package p\nfunc A() int { return 1 }\nfunc B() int { return 2 }\n")
	generated := map[string][]byte{"a.go": []byte("package p\nfunc A() int { return 1 }\n"), "b.go": []byte("package p\nfunc B() int { return 2 }\n")}
	if _, err := callbackExtractionPartitionDigest(intermediate, generated); err != nil {
		t.Fatal(err)
	}
	generated["b.go"] = []byte("package p\nfunc B() int { return 3 }\n")
	var failure Failure
	_, err := callbackExtractionPartitionDigest(intermediate, generated)
	if !errors.As(err, &failure) || failure.Reason != "CALLBACK_DECLARATION_PARTITION_REFUTED" {
		t.Fatalf("mutated final unit was accepted: %v", err)
	}
}
