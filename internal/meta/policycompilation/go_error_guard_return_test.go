package policycompilation

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

//go:embed testdata/go-error-guard/profile-original.go.golden
var goReturnGuardOriginal []byte

//go:embed testdata/go-error-guard/profile-delivery-oracle.go.golden
var goReturnGuardOracle []byte

const goReturnGuardHandler = "{\n\t\t\treturn exitFailure\n\t\t}"

func goReturnGuardFixture(source []byte, function string) []byte {
	return fmt.Appendf(nil, "package goerrorguard\nnamespace goerrorguard\n"+
		"entity Source id \"gooo://error-guard/source\"\n"+
		"entity Candidate id \"gooo://error-guard/candidate\"\n"+
		"activity GuardWrite(Source) -> Candidate computes \"go-error-guard:v2;function=%s;writer=stdout;writer-type=interfaceWriter;source=%s;handler=%s\"\n",
		function, DigestBytes(source), DigestBytes([]byte(goReturnGuardHandler)))
}

func TestGoReturnGuardProposesFromActualPublicProfileSources(t *testing.T) {
	for _, function := range []string{"runMetaPolicyGenerationProfile", "runMetaPolicyRevisionProfile"} {
		t.Run(function, func(t *testing.T) {
			program := goReturnGuardFixture(goReturnGuardOriginal, function)
			proposal, err := ProposeGoErrorGuard("profile-guard.gooo", program, goReturnGuardOriginal)
			if err != nil || proposal.State != "PROPOSED" {
				t.Fatalf("actual profile proposal: %+v error=%v", proposal, err)
			}
			call := goReturnGuardOriginal[proposal.EditStart:proposal.EditEnd]
			want := string(goReturnGuardOriginal[:proposal.EditStart]) +
				fmt.Sprintf("if _, err := %s; err != nil %s", call, goReturnGuardHandler) +
				string(goReturnGuardOriginal[proposal.EditEnd:])
			if proposal.CandidateSource != want || proposal.HandlerDigest != DigestBytes([]byte(goReturnGuardHandler)) ||
				proposal.ProgramDigest != DigestBytes(program) || proposal.SemanticDigest == "" ||
				proposal.MutationAuthority != 0 || proposal.PromotionAuthority != 0 || proposal.Improvement != "UNKNOWN" {
				t.Fatalf("proposal changed its bounded contract: %+v", proposal)
			}
			payload, err := json.Marshal(proposal)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("gooo return guard proposal-json: %s", payload)
		})
	}
}

func TestGoReturnGuardUsesDeclaredWriterType(t *testing.T) {
	source := []byte(strings.ReplaceAll(string(goReturnGuardOriginal), "interfaceWriter", "outputWriter"))
	program := []byte(strings.ReplaceAll(string(goReturnGuardFixture(source, "runMetaPolicyGenerationProfile")),
		"writer-type=interfaceWriter", "writer-type=outputWriter"))
	report, err := ProposeGoErrorGuard("renamed.guard.gooo", program, source)
	if err != nil || report.State != "PROPOSED" {
		t.Fatalf("declared writer type was not used: %+v error=%v", report, err)
	}
}

func TestGoReturnGuardUnsupportedContextsStayUnknown(t *testing.T) {
	original := string(goReturnGuardOriginal)
	tests := []struct {
		name   string
		source string
	}{
		{"missing-interface", strings.Replace(original, "type interfaceWriter interface", "type otherWriter interface", 1)},
		{"wrong-method", strings.Replace(original, "Write([]byte) (int, error)", "Write([]byte) error", 1)},
		{"extra-method", strings.Replace(original, "Write([]byte) (int, error)", "Write([]byte) (int, error)\n\tFlush() error", 1)},
		{"alias", strings.Replace(original, "type interfaceWriter interface", "type interfaceWriter = interface", 1)},
		{"handler-effect", strings.Replace(original, goReturnGuardHandler, "{\n\t\t\tfmt.Fprintln(stderr, err)\n\t\t\treturn exitFailure\n\t\t}", 1)},
		{"wrong-writer", strings.Replace(original, "fmt.Fprintf(stdout, \"generated profile:", "fmt.Fprintf(stderr, \"generated profile:", 1)},
		{"shadowed-import", strings.Replace(original, "\tif jsonMode {\n", "\tfmt := struct{}{}\n\tif jsonMode {\n", 1)},
		{"ambiguous", original + "\nfunc runMetaPolicyGenerationProfile() int { return 0 }\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(test.source)
			program := goReturnGuardFixture(source, "runMetaPolicyGenerationProfile")
			report, err := ProposeGoErrorGuard("unsupported.guard.gooo", program, source)
			if err == nil || report.State != "UNKNOWN" || report.Pending == nil ||
				report.Pending.Stage == "" || report.Pending.Step == "" || report.Pending.Reason == "" ||
				report.Pending.UnknownClass == "" || report.Pending.NextOperation == "" ||
				report.Pending.BlockedBy == nil || report.CandidateSource != "" {
				t.Fatalf("unsupported context was not explicit: %+v error=%v", report, err)
			}
		})
	}
}

func TestGoReturnGuardPinsDoNotRepairThemselves(t *testing.T) {
	program := goReturnGuardFixture(goReturnGuardOriginal, "runMetaPolicyGenerationProfile")
	staleSource := append([]byte(string(goReturnGuardOriginal)), '\n')
	report, err := ProposeGoErrorGuard("stale.guard.gooo", program, staleSource)
	if err == nil || report.State != "REFUTED" || report.Reason != "GO_SOURCE_PIN_MISMATCH" || report.CandidateSource != "" {
		t.Fatalf("stale source was accepted: %+v error=%v", report, err)
	}
	wrongHandler := []byte(strings.Replace(string(program), DigestBytes([]byte(goReturnGuardHandler)), DigestBytes([]byte("different handler")), 1))
	report, err = ProposeGoErrorGuard("stale-handler.guard.gooo", wrongHandler, goReturnGuardOriginal)
	if err == nil || report.State != "REFUTED" || report.Reason != "GO_HANDLER_PIN_MISMATCH" || report.CandidateSource != "" {
		t.Fatalf("stale handler was accepted: %+v error=%v", report, err)
	}
}
