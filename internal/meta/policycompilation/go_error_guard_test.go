package policycompilation

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

// This is the actual pre-#849 file at a97a102, not a fabricated bad revision.
//
//go:embed testdata/go-error-guard/original.go.golden
var goErrorGuardOriginal []byte

const goErrorGuardHandler = "{\n\t\t\tfmt.Fprintf(stderr, \"gooo: run package: %v\\n\", err)\n\t\t\treturn exitFailure\n\t\t}"

func goErrorGuardFixtureProgram(source []byte, function, writer, diagnostic, handler string) []byte {
	program := fmt.Sprintf("go-error-guard:v1;function=%s;writer=%s;diagnostic=%s;source=%s;handler=%s",
		function, writer, diagnostic, DigestBytes(source), DigestBytes([]byte(handler)))
	return []byte(fmt.Sprintf("package goerrorguard\nnamespace goerrorguard\n"+
		"entity Source id \"gooo://error-guard/source\"\n"+
		"entity Candidate id \"gooo://error-guard/candidate\"\n"+
		"activity GuardWrite(Source) -> Candidate computes %q\n", program))
}

func goErrorGuardFixture() []byte {
	return goErrorGuardFixtureProgram(goErrorGuardOriginal, "writeSourcePackageResult", "stdout", "stderr", goErrorGuardHandler)
}

func TestGoErrorGuardProposalPreservesPinnedSourceAndAuthority(t *testing.T) {
	program := goErrorGuardFixture()
	original := bytes.Clone(goErrorGuardOriginal)
	first, err := ProposeGoErrorGuard("guard.gooo", program, original)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ProposeGoErrorGuard("guard.gooo", program, original)
	if err != nil || first.CandidateSource != replay.CandidateSource || first.CandidateDigest != replay.CandidateDigest {
		t.Fatalf("candidate replay differs: %v", err)
	}
	if first.State != "PROPOSED" || first.ActivityID == "" || first.SemanticDigest == "" ||
		first.CandidateDigest == first.SourceDigest || first.ProgramDigest != DigestBytes(program) ||
		first.Admission.State != "UNKNOWN" || first.MutationAuthority != 0 || first.PromotionAuthority != 0 ||
		first.Improvement != "UNKNOWN" || !bytes.Equal(original, goErrorGuardOriginal) {
		t.Fatalf("source or proposal authority changed: %#v", first)
	}
	if !strings.HasPrefix(first.CandidateSource, string(original[:first.EditStart])) ||
		!strings.HasSuffix(first.CandidateSource, string(original[first.EditEnd:])) ||
		strings.Count(first.CandidateSource, goErrorGuardHandler) != 2 {
		t.Fatal("guard changed bytes outside the selected call or did not copy the declared handler")
	}
}

func TestGoErrorGuardUnknownAndRefutedNeverProduceCandidates(t *testing.T) {
	unsupported := bytes.Replace(goErrorGuardOriginal, []byte("fmt.Fprintf(stdout,"), []byte("fmt.Fprint(stdout,"), 1)
	start := bytes.Index(goErrorGuardOriginal, []byte("func writeSourcePackageResult("))
	end := bytes.Index(goErrorGuardOriginal, []byte("\nfunc packageDirectoryArgument("))
	ambiguous := append(bytes.Clone(goErrorGuardOriginal), goErrorGuardOriginal[start:end]...)
	shadowed := bytes.Replace(goErrorGuardOriginal, []byte("failureCode int) int"), []byte("failureCode int, fmt any) int"), 1)
	cases := []struct {
		name, state, class string
		program, source    []byte
	}{
		{"invalid-gooo", "REFUTED", "", []byte("not a program"), goErrorGuardOriginal},
		{"unknown-profile", "REFUTED", "", bytes.Replace(goErrorGuardFixture(), []byte("guard:v1"), []byte("guard:v99"), 1), goErrorGuardOriginal},
		{"stale-source", "REFUTED", "", goErrorGuardFixture(), append(bytes.Clone(goErrorGuardOriginal), '\n')},
		{"stale-handler", "REFUTED", "", goErrorGuardFixtureProgram(goErrorGuardOriginal, "writeSourcePackageResult", "stdout", "stderr", "{}"), goErrorGuardOriginal},
		{"unsupported-call", "UNKNOWN", "UNBOUNDED", goErrorGuardFixtureProgram(unsupported, "writeSourcePackageResult", "stdout", "stderr", goErrorGuardHandler), unsupported},
		{"ambiguous-function", "UNKNOWN", "AMBIGUOUS", goErrorGuardFixtureProgram(ambiguous, "writeSourcePackageResult", "stdout", "stderr", goErrorGuardHandler), ambiguous},
		{"shadowed-import", "UNKNOWN", "UNBOUNDED", goErrorGuardFixtureProgram(shadowed, "writeSourcePackageResult", "stdout", "stderr", goErrorGuardHandler), shadowed},
		{"missing-function", "UNKNOWN", "DIRECT_MISSING", goErrorGuardFixtureProgram(goErrorGuardOriginal, "missing", "stdout", "stderr", goErrorGuardHandler), goErrorGuardOriginal},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			report, err := ProposeGoErrorGuard("guard.gooo", item.program, item.source)
			if err == nil || report.State != item.state || report.CandidateSource != "" || report.CandidateDigest != "" ||
				report.MutationAuthority != 0 || report.PromotionAuthority != 0 || report.Admission.State != "UNKNOWN" {
				t.Fatalf("invalid input produced a candidate or authority: %#v err=%v", report, err)
			}
			if item.class != "" && (report.Pending == nil || report.Pending.Stage == "" || report.Pending.Step == "" ||
				report.Pending.Reason == "" || report.Pending.UnknownClass != item.class ||
				report.Pending.NextOperation == "" || report.Pending.BlockedBy == nil) {
				t.Fatal("UNKNOWN lost its causal fields")
			}
		})
	}
}

func TestGoErrorGuardSelectorsComeFromGooo(t *testing.T) {
	rename := strings.NewReplacer("stdout", "destination", "stderr", "diagnostics", "writeSourcePackageResult", "writeOtherResult")
	source := []byte(rename.Replace(string(goErrorGuardOriginal)))
	handler := rename.Replace(goErrorGuardHandler)
	program := goErrorGuardFixtureProgram(source, "writeOtherResult", "destination", "diagnostics", handler)
	report, err := ProposeGoErrorGuard("renamed.gooo", program, source)
	if err != nil || report.State != "PROPOSED" ||
		!strings.Contains(report.CandidateSource, "if _, err := fmt.Fprintf(destination,") {
		t.Fatalf("host ignored source-defined selectors: %#v err=%v", report, err)
	}
	stale, err := ProposeGoErrorGuard("renamed.gooo", program, []byte(report.CandidateSource))
	if err == nil || stale.State != "REFUTED" || stale.Reason != "GO_SOURCE_PIN_MISMATCH" {
		t.Fatal("changed source was mistaken for a fixed point or silently repinned")
	}
}
