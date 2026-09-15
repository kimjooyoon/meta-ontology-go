package policycompilation

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func goGuardPipelineFixture(source []byte, function string) []byte {
	return fmt.Appendf(nil, "package goerrorguard\nnamespace goerrorguard\n"+
		"entity Source id \"gooo://error-guard/source\"\n"+
		"entity RawCandidate id \"gooo://error-guard/raw-candidate\"\n"+
		"entity Candidate id \"gooo://error-guard/candidate\"\n"+
		"activity GuardWrite(Source) -> RawCandidate computes \"go-error-guard:v2;function=%s;writer=stdout;writer-type=interfaceWriter;source=%s;handler=%s\"\n"+
		"activity Canonicalize(RawCandidate) -> Candidate computes \"go-source-format:v1;toolchain=%s\"\n",
		function, DigestBytes(source), DigestBytes([]byte(goReturnGuardHandler)), runtime.Version())
}

func TestGoGuardPipelinePreservesRawProposalAndIndependentAdmission(t *testing.T) {
	source := bytes.Clone(goReturnGuardOriginal)
	program := goGuardPipelineFixture(source, "runMetaPolicyGenerationProfile")
	report, err := ProposeGoErrorGuardPipeline("pipeline.gooo", program, source)
	if err != nil || report.State != "PROPOSED" || report.Guard == nil || report.Rendering == nil {
		t.Fatalf("pipeline did not propose: %+v error=%v", report, err)
	}
	raw, rendered := report.Guard, report.Rendering
	if raw.State != "PROPOSED" || rendered.State != "RENDERED" || rendered.NativeCalls != 3 ||
		raw.CandidateSource == report.CandidateSource || raw.CandidateDigest != rendered.InputDigest ||
		report.CandidateDigest != rendered.OutputDigest || rendered.ReplayDigest != rendered.OutputDigest ||
		!rendered.FixedPoint || !rendered.OutsideEditPreserved || rendered.Toolchain != runtime.Version() {
		t.Fatalf("raw and canonical identities were not retained: %+v", report)
	}
	if report.ProgramDigest != DigestBytes(program) || raw.ProgramDigest != report.ProgramDigest ||
		report.SemanticDigest == "" || raw.SemanticDigest != report.SemanticDigest ||
		raw.ActivityID == "" || rendered.ActivityID == "" || raw.ActivityID == rendered.ActivityID ||
		report.Admission.State != "UNKNOWN" || raw.Admission.State != "UNKNOWN" ||
		report.MutationAuthority != 0 || report.PromotionAuthority != 0 || report.Improvement != "UNKNOWN" ||
		!bytes.Equal(source, goReturnGuardOriginal) {
		t.Fatalf("pipeline invented authority or changed input: %+v", report)
	}
	replay, err := ProposeGoErrorGuardPipeline("pipeline.gooo", program, source)
	if err != nil || replay.CandidateSource != report.CandidateSource || replay.CandidateDigest != report.CandidateDigest {
		t.Fatalf("pipeline replay differs: %+v error=%v", replay, err)
	}
	legacy, err := ProposeGoErrorGuard("pipeline.gooo", program, source)
	if err == nil || legacy.State != "REFUTED" || legacy.CandidateSource != "" {
		t.Fatal("single-activity API silently accepted the pipeline")
	}
}

func TestGoGuardPipelineInvalidProgramsDoNotRunRenderer(t *testing.T) {
	program := string(goGuardPipelineFixture(goReturnGuardOriginal, "runMetaPolicyGenerationProfile"))
	cases := []struct {
		name    string
		program string
	}{
		{"disconnected", strings.Replace(program, "Canonicalize(RawCandidate)", "Canonicalize(Source)", 1)},
		{"extra-node", program + "entity Extra id \"gooo://error-guard/extra\"\n"},
		{"unsupported-renderer", strings.Replace(program, "go-source-format:v1", "go-source-format:v99", 1)},
		{"implicit-toolchain", strings.Replace(program, ";toolchain="+runtime.Version(), "", 1)},
		{"extra-renderer-field", strings.Replace(program, "toolchain="+runtime.Version(), "toolchain="+runtime.Version()+";admission=FIXED_POINT", 1)},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			report, err := ProposeGoErrorGuardPipeline("invalid.gooo", []byte(test.program), goReturnGuardOriginal)
			if err == nil || report.State != "REFUTED" || report.Guard != nil ||
				report.Rendering != nil || report.CandidateSource != "" || report.CandidateDigest != "" {
				t.Fatalf("invalid pipeline produced output: %+v error=%v", report, err)
			}
		})
	}
}

func TestGoGuardPipelineRetainsDirectAndBlockedUnknown(t *testing.T) {
	program := goGuardPipelineFixture(goReturnGuardOriginal, "MissingFunction")
	report, err := ProposeGoErrorGuardPipeline("missing.gooo", program, goReturnGuardOriginal)
	if err == nil || report.State != "UNKNOWN" || report.Guard == nil || report.Rendering == nil {
		t.Fatalf("missing guard was not retained: %+v error=%v", report, err)
	}
	direct, blocked := report.Guard.Pending, report.Rendering.Pending
	requireGoGuardPipelinePending(t, direct, "DIRECT_MISSING")
	requireGoGuardPipelinePending(t, blocked, "DEPENDENCY_BLOCKED")
	if len(direct.BlockedBy) != 0 || len(blocked.BlockedBy) != 1 ||
		blocked.BlockedBy[0] != report.Guard.ActivityID || report.Rendering.NativeCalls != 0 ||
		report.CandidateSource != "" || report.CandidateDigest != "" {
		t.Fatalf("renderer guessed through missing upstream: %+v", report)
	}
}

func requireGoGuardPipelinePending(t *testing.T, pending *PolicyRevisionPending, class string) {
	t.Helper()
	if pending == nil || pending.State != "UNKNOWN" || pending.Stage == "" ||
		pending.Step == "" || pending.Reason == "" || pending.UnknownClass != class ||
		pending.NextOperation == "" || pending.BlockedBy == nil {
		t.Fatalf("UNKNOWN lost its causal frontier: %+v", pending)
	}
}

func TestGoGuardPipelinePinContradictionsStayRefuted(t *testing.T) {
	program := goGuardPipelineFixture(goReturnGuardOriginal, "runMetaPolicyGenerationProfile")
	stale := append(bytes.Clone(goReturnGuardOriginal), '\n')
	report, err := ProposeGoErrorGuardPipeline("stale.gooo", program, stale)
	if err == nil || report.State != "REFUTED" || report.Reason != "GO_SOURCE_PIN_MISMATCH" ||
		report.Rendering == nil || report.Rendering.NativeCalls != 0 || report.CandidateSource != "" {
		t.Fatalf("source contradiction was hidden: %+v error=%v", report, err)
	}
	version := []byte(strings.Replace(string(program), "toolchain="+runtime.Version(), "toolchain=go1.0.0", 1))
	report, err = ProposeGoErrorGuardPipeline("toolchain.gooo", version, goReturnGuardOriginal)
	if err == nil || report.State != "REFUTED" || report.Reason != "GO_RENDER_TOOLCHAIN_PIN_MISMATCH" ||
		report.Rendering == nil || report.Rendering.NativeCalls != 0 || report.CandidateSource != "" {
		t.Fatalf("toolchain contradiction was hidden: %+v error=%v", report, err)
	}
}

func TestGoGuardPipelineRefusesUnrelatedFormatting(t *testing.T) {
	source := append(bytes.Clone(goReturnGuardOriginal), '\n')
	program := goGuardPipelineFixture(source, "runMetaPolicyGenerationProfile")
	report, err := ProposeGoErrorGuardPipeline("noncanonical.gooo", program, source)
	if err == nil || report.State != "UNKNOWN" || report.Reason != "GO_RENDER_ORIGINAL_NOT_CANONICAL" ||
		report.Guard == nil || report.Guard.State != "PROPOSED" || report.Rendering == nil ||
		report.Rendering.NativeCalls != 1 || report.CandidateSource != "" {
		t.Fatalf("pipeline formatted unrelated input: %+v error=%v", report, err)
	}
	requireGoGuardPipelinePending(t, report.Rendering.Pending, "UNBOUNDED")
	if goGuardCanonicalSpan([]byte("before-call-after"), []byte("changed-call-after"), 7, 11) ||
		goGuardCanonicalSpan([]byte("before-call-after"), []byte("before-call-after"), -1, 11) {
		t.Fatal("outside-span renderer guard accepted a changed prefix or invalid coordinate")
	}
}
