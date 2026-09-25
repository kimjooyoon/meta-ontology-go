package extractor

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPaginationCallbackFactoryOriginalAndGeneratedCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("callback factory execution is CI-only")
	}
	root := callbackPreviewRepositoryRoot(t)
	preview, err := PreviewBoundedPaginationCallbackWithStrategy(root, callbackPreviewLogicalPath, callbackPreviewFactoryLowering)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCallbackPreviewResult(preview); err != nil {
		t.Fatal(err)
	}
	if preview.Candidate == nil || preview.State != "UNKNOWN" || preview.Candidate.Promotion != "NONE" ||
		preview.OperationResultAdmission != "FORBIDDEN" || preview.ApplyPermission != "FORBIDDEN" {
		t.Fatalf("factory preview escaped admission boundary: %+v", preview)
	}
	if len(preview.Captures) != 2 || !strings.Contains(preview.Candidate.HelperSource, "*atomic.Int32") ||
		!strings.Contains(preview.Candidate.HelperSource, "return func(") || strings.HasPrefix(preview.Candidate.WrapperSource, "func(") {
		t.Fatalf("factory did not retain the callback and capture addresses: %+v", preview.Candidate)
	}
	helper, err := renderedFunctionHelper([]byte(preview.Candidate.CandidateSource), preview.Candidate.HelperName)
	if err != nil || physicalLines(helper) > functionLineLimit {
		t.Fatalf("factory rendered helper lines=%d err=%v", physicalLines(helper), err)
	}
	typeCheckCallbackPreviewCandidate(t, root, callbackPreviewLogicalPath, preview.Candidate.CandidateSource)
	runCallbackPreviewFixtureSuite(t, root, callbackPreviewLogicalPath, "")
	runCallbackPreviewFixtureSuite(t, root, callbackPreviewLogicalPath, preview.Candidate.CandidateSource)
	t.Logf("factory original/generated suites=2 captures=%d pending_effects=%d resolved_effects=%d helper_lines=%d rendered_helper_lines=%d parent_lines=%d apply=%s",
		len(preview.Captures), len(preview.PendingEffects), preview.Evidence.ResolvedEffectCount,
		preview.Candidate.HelperLines, physicalLines(helper), preview.Candidate.ParentFunctionLines, preview.ApplyPermission)
	raw, err := json.Marshal(preview)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CallbackPreviewResult
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCallbackPreviewResult(decoded); err != nil {
		t.Fatal(err)
	}
	decoded.LoweringStrategy = callbackPreviewWrapperLowering
	if err := ValidateCallbackPreviewResult(decoded); err == nil {
		t.Fatal("factory output accepted a different native lowering strategy")
	}
}

func TestCallbackFactorySharedBindingsShadowingAndRecoverCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("callback factory execution is CI-only")
	}
	cases := []struct {
		name string
		body string
	}{
		{"shared-mutation", `value := 1
callback := h.HandlerFunc(func(_ h.ResponseWriter, _ *h.Request) { value++ })
value = 4
callback(nil, nil)
if value != 5 { t.Fatalf("shared value=%d", value) }`},
		{"typed-shadowing-and-hygiene", `value := 2
callback := h.HandlerFunc(func(_ h.ResponseWriter, _ *h.Request) {
value++
{ value := 40; value++; if value != 41 { panic("shadow changed") } }
goooCapture0 := 9
if goooCapture0 != 9 { panic("hygiene changed") }
value++
})
callback(nil, nil)
if value != 4 { t.Fatalf("outer value=%d", value) }`},
		{"direct-deferred-recover", `recovered := false
callback := h.HandlerFunc(func(_ h.ResponseWriter, _ *h.Request) {
if recover() != nil { recovered = true }
})
func() { defer callback(nil, nil); panic("expected") }()
if !recovered { t.Fatal("deferred callback lost direct recover") }`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := "package fixture\nimport h \"net/http\"\nimport \"testing\"\nfunc " + callbackPreviewTarget + "(t *testing.T) {\n" + testCase.body + "\n}\n"
			root := writeCallbackFactoryFixture(t, source)
			preview, err := PreviewBoundedPaginationCallbackWithStrategy(root, "fixture_test.go", callbackPreviewFactoryLowering)
			if err != nil || preview.Candidate == nil {
				t.Fatalf("preview=%+v err=%v", preview, err)
			}
			if err := ValidateCallbackPreviewResult(preview); err != nil {
				t.Fatal(err)
			}
			runCallbackFactoryFixture(t, root)
			runCallbackFactoryFixture(t, writeCallbackFactoryFixture(t, preview.Candidate.CandidateSource))
		})
	}
}

func TestCallbackFactoryRejectsUnknownLowering(t *testing.T) {
	for _, lowering := range []string{"", "FIXED_POINT", "apply"} {
		if _, err := PreviewBoundedPaginationCallbackWithStrategy("", "", lowering); err == nil || !strings.Contains(err.Error(), "unsupported callback preview lowering") {
			t.Fatalf("lowering=%q err=%v, want rejection before reading input", lowering, err)
		}
	}
}

func writeCallbackFactoryFixture(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module factory.test\n\ngo 1.27.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func runCallbackFactoryFixture(t *testing.T, root string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-count=1", "-run", "^"+callbackPreviewTarget+"$", ".")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("callback factory runtime fixture: %v\n%s", err, output)
	}
}
