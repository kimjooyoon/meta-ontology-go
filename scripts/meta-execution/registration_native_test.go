package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

const registrationNestedEnvironment = "GOOO_REGISTRATION_NATIVE_NESTED"

func TestNativeRegistrationUsesCommonTwoOperationExecution(t *testing.T) {
	if os.Getenv("CI") != "true" || os.Getenv("GOOO_REGISTRATION_NATIVE_E2E") != "1" ||
		os.Getenv(registrationNestedEnvironment) == "1" {
		t.Skip("real common registration execution belongs to dedicated native Actions")
	}
	t.Setenv(registrationNestedEnvironment, "1")
	t.Setenv(collapseNestedE2EEnvironment, "1")
	root := collapseTestRepositoryRoot(t)
	head := collapseTestHead(t, root)
	temporary := t.TempDir()
	snapshot := filepath.Join(temporary, "input")
	registrationTestCommand(t, root, "git", "clone", "--no-hardlinks", "--quiet", root, snapshot)
	registrationTestCommand(t, snapshot, "git", "checkout", "--quiet", "--detach", head)
	if err := restoreCollapseFixture(snapshot); err != nil {
		t.Fatal(err)
	}
	request := registrationTestRequest(t, snapshot)
	binary := filepath.Join(temporary, "meta-execution")
	buildStarted := time.Now()
	registrationTestCommand(t, root, "go", "build", "-mod=readonly", "-o", binary, "./scripts/meta-execution")
	buildMS := time.Since(buildStarted).Milliseconds()
	requestPath := filepath.Join(temporary, "request.json")
	registrationTestWriteJSON(t, requestPath, request)
	pinned := registrationTestCommand(t, snapshot, binary, "--registration-mode=inspect",
		"--registration-root="+snapshot, "--registration-request="+requestPath)
	var err error
	request, err = syntaxregistration.DecodeRequest(pinned)
	if err != nil {
		t.Fatal(err)
	}
	registrationTestWriteJSON(t, requestPath, request)
	subject := sourcepolicy.SourceSubject{Path: "scripts/meta-execution/collapsefixture/collapse_fixture.go",
		Line: 3, Name: "CollapseFixture"}
	_, collapseAction, sourceReport := collapsePlannerFixture(t, subject.String(), head)
	sourceReport.Indicators = []sourcepolicy.Indicator{collapseAction.SourceIndicator}
	metricsPath := filepath.Join(temporary, "source-metrics.json")
	registrationTestWriteJSON(t, metricsPath, linecaps.LineMetricsReport{
		Root: snapshot, CommitSHA: head, Meta: sourceReport})
	planBytes := registrationTestCommand(t, snapshot, binary, "--registration-mode=plan",
		"--registration-root="+snapshot, "--registration-request="+requestPath,
		"--registration-base="+head, "--source-metrics="+metricsPath)
	var plan generation.Plan
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Decision != generation.DecisionPlan || len(plan.Selected) != 2 ||
		plan.MinimumIndependent != 2 || plan.RequestedK != 2 || plan.PromotionAuthorized {
		t.Fatalf("real common plan is not exact: %+v", plan)
	}
	planPath, manifestPath := filepath.Join(temporary, "plan.json"), filepath.Join(temporary, "manifest.json")
	registrationTestWriteJSON(t, planPath, plan)
	beforeStatus := registrationTestCommand(t, snapshot, "git", "--no-optional-locks",
		"status", "--porcelain=v1", "--untracked-files=all")
	executionStarted := time.Now()
	output := registrationTestCommand(t, snapshot, binary, "--plan="+planPath,
		"--output="+manifestPath, "--source-metrics="+metricsPath)
	executionMS := time.Since(executionStarted).Milliseconds()
	var manifest generation.ExecutionManifest
	var bundle generation.OperationObservationBundle
	registrationTestReadJSON(t, manifestPath, &manifest)
	registrationTestReadJSON(t, filepath.Join(temporary, "meta-operation-observations.json"), &bundle)
	if err := generation.ValidateObservationBundle(bundle, plan, manifest); err != nil {
		t.Fatalf("actual common bundle failed: %v\n%s", err, output)
	}
	receipts := generation.VerifyReceiptsWithFailures(plan, bundle.Receipts, bundle.Failures)
	if receipts.Decision != generation.ReceiptDecisionConformant || len(bundle.Receipts) != 2 ||
		len(bundle.Failures) != 0 || receipts.PromotionAuthorized {
		t.Fatalf("actual common receipts are not conformant: %+v\n%s", receipts, output)
	}
	registrationAssertNativeReceipts(t, plan, bundle, request)
	afterStatus := registrationTestCommand(t, snapshot, "git", "--no-optional-locks",
		"status", "--porcelain=v1", "--untracked-files=all")
	repinned := registrationTestCommand(t, snapshot, binary, "--registration-mode=inspect",
		"--registration-root="+snapshot, "--registration-request="+requestPath)
	if !bytes.Equal(beforeStatus, afterStatus) || !bytes.Equal(pinned, repinned) {
		t.Fatal("common executor changed its input project or pinned source view")
	}
	candidateBytes := registrationTestCommand(t, snapshot, binary, "--registration-mode=worker",
		"--registration-root="+snapshot, "--registration-request="+requestPath)
	var candidate syntaxregistration.Candidate
	if err := json.Unmarshal(candidateBytes, &candidate); err != nil {
		t.Fatal(err)
	}
	if candidate.RequestDigest != syntaxregistration.RequestDigest(request) || candidate.ApplyAuthorized ||
		candidate.PromotionAllowed || candidate.RequiredArtifacts != 9 || len(candidate.Artifacts) != 9 {
		t.Fatal("evidence capture lost exact request or non-authorizing boundary")
	}
	registrationPublishNativeEvidence(t, plan, manifest, bundle, receipts, candidate,
		buildMS, executionMS, output)
}
