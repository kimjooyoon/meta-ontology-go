package extractor

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestClosurePreservationReplaysAgainstOriginalSourceCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("closure preservation source replay is CI-only")
	}
	root := callbackPreviewRepositoryRoot(t)
	preview, err := PreviewBoundedPaginationCallbackWithStrategy(root, callbackPreviewLogicalPath, callbackPreviewFactoryLowering)
	if err != nil || preview.Candidate == nil || preview.StructureProof == nil {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	proof, err := VerifyCallbackFactoryPreservation(root, callbackPreviewLogicalPath, *preview.Candidate)
	if err != nil {
		t.Fatal(err)
	}
	if proof.ProofDigest != preview.StructureProof.ProofDigest || len(proof.Captures) != 2 || proof.CaptureReferences <= 0 || proof.CallExpressions <= 0 || proof.SemanticAdmission != "UNKNOWN" {
		t.Fatalf("replayed proof=%+v", proof)
	}
	if err := ValidateCallbackPreviewResult(preview); err != nil {
		t.Fatal(err)
	}
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
	decoded.StructureProof.SemanticAdmission = "CLOSED"
	decoded.StructureProof.ProofDigest = closurePreservationDigest(*decoded.StructureProof)
	if err := ValidateCallbackPreviewResult(decoded); err == nil {
		t.Fatal("structural proof promoted semantic admission")
	}
	t.Logf("source-bound proof captures=%d references=%d AST_call_expressions=%d body_pairs=1/1 context_pairs=1/1 semantic_admission=%s",
		len(proof.Captures), proof.CaptureReferences, proof.CallExpressions, proof.SemanticAdmission)
}

func TestClosurePreservationRejectsCoherentSourceMutantsCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("closure preservation mutants are CI-only")
	}
	root := callbackPreviewRepositoryRoot(t)
	preview, err := PreviewBoundedPaginationCallbackWithStrategy(root, callbackPreviewLogicalPath, callbackPreviewFactoryLowering)
	if err != nil || preview.Candidate == nil {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	cases := []struct {
		name   string
		mutate func(*CallbackPreviewCandidate)
	}{
		{"callback-body", func(c *CallbackPreviewCandidate) { mutateClosureHelper(t, c, ".Add(1)", ".Add(2)") }},
		{"copied-capture", func(c *CallbackPreviewCandidate) {
			mutateClosureHelper(t, c, "*paginationFixture", "paginationFixture")
			mutateClosureHelper(t, c, "(*goooCapture0)", "goooCapture0")
			c.CandidateSource = strings.Replace(c.CandidateSource, c.WrapperSource, strings.Replace(c.WrapperSource, "&fixture", "fixture", 1), 1)
			c.WrapperSource = strings.Replace(c.WrapperSource, "&fixture", "fixture", 1)
		}},
		{"wrong-capture-address", func(c *CallbackPreviewCandidate) {
			c.CandidateSource = strings.Replace(c.CandidateSource, c.WrapperSource, strings.Replace(c.WrapperSource, "&fixture", "&fixtures", 1), 1)
			c.WrapperSource = strings.Replace(c.WrapperSource, "&fixture", "&fixtures", 1)
		}},
		{"constructor-effect", func(c *CallbackPreviewCandidate) {
			mutateClosureHelper(t, c, "return func(", "println(\"extra\")\nreturn func(")
		}},
		{"unrelated-source", func(c *CallbackPreviewCandidate) {
			c.CandidateSource = strings.Replace(c.CandidateSource, "runtime caller unavailable", "tampered caller", 1)
		}},
		{"stale-source-digest", func(c *CallbackPreviewCandidate) { c.SourceDigest = callbackPreviewDigest([]byte("not the source")) }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			candidate := *preview.Candidate
			testCase.mutate(&candidate)
			if candidate.CandidateSource == preview.Candidate.CandidateSource && candidate.SourceDigest == preview.Candidate.SourceDigest {
				t.Fatal("counterexample did not change its input")
			}
			candidate.CandidateDigest = callbackPreviewDigest([]byte(candidate.CandidateSource))
			candidate.HelperBytes, candidate.HelperLines = len(candidate.HelperSource), physicalLines([]byte(candidate.HelperSource))
			if _, err := VerifyCallbackFactoryPreservation(root, callbackPreviewLogicalPath, candidate); err == nil {
				t.Fatal("source-bound proof accepted a coherently rehashed counterexample")
			}
		})
	}
}

func mutateClosureHelper(t *testing.T, candidate *CallbackPreviewCandidate, old, replacement string) {
	t.Helper()
	helper := strings.ReplaceAll(candidate.HelperSource, old, replacement)
	if helper == candidate.HelperSource || !strings.Contains(candidate.CandidateSource, candidate.HelperSource) {
		t.Fatal("helper counterexample has no exact source target")
	}
	candidate.CandidateSource = strings.Replace(candidate.CandidateSource, candidate.HelperSource, helper, 1)
	candidate.HelperSource = helper
}

func TestClosureStructureDoesNotProveCallerIdentityCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("caller-observability counterexample is CI-only")
	}
	source := `package fixture
import h "net/http"
import "runtime"
import "testing"
func TestPaginationFixturesExecuteParserAndHTTPClient(t *testing.T) {
name := ""
handler := h.HandlerFunc(func(_ h.ResponseWriter, _ *h.Request) {
pc, _, _, ok := runtime.Caller(0)
if ok { name = runtime.FuncForPC(pc).Name() }
})
handler(nil, nil)
if name == "" { t.Fatal("missing caller identity") }
t.Log("OBSERVED_CALLER=" + name)
}
`
	root := writeCallbackFactoryFixture(t, source)
	preview, err := PreviewBoundedPaginationCallbackWithStrategy(root, "fixture_test.go", callbackPreviewFactoryLowering)
	if err != nil || preview.StructureProof == nil || preview.Candidate == nil {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	before := observedClosureCaller(t, root)
	after := observedClosureCaller(t, writeCallbackFactoryFixture(t, preview.Candidate.CandidateSource))
	if before == after || preview.StructureProof.SemanticAdmission != "UNKNOWN" || preview.ApplyPermission != "FORBIDDEN" {
		t.Fatalf("caller before=%q after=%q admission=%s apply=%s", before, after, preview.StructureProof.SemanticAdmission, preview.ApplyPermission)
	}
	t.Logf("caller-identity counterexample before=%s after=%s structural_state=%s semantic_admission=%s",
		before, after, preview.StructureProof.State, preview.StructureProof.SemanticAdmission)
}

func observedClosureCaller(t *testing.T, root string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-count=1", "-v", "-run", "^"+callbackPreviewTarget+"$", ".")
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("caller observation: %v\n%s", err, output)
	}
	for line := range strings.SplitSeq(string(output), "\n") {
		if _, value, found := strings.Cut(line, "OBSERVED_CALLER="); found && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	t.Fatalf("caller observation was not emitted: %s", output)
	return ""
}
