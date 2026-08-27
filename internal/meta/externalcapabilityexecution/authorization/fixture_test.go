package authorization

import (
	"strings"
	"testing"

	capability "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution"
)

func fixtureInput() Input {
	subject := strings.Repeat("a", 40)
	report := capability.Report{Schema: capability.ReportSchema, SubjectSHA: subject,
		Decision: capability.DecisionExecutable, Resolution: capability.ResolutionExact,
		Completed: 10, Total: 10, BasisPoints: 10000, RepositoryWrites: 0,
		ExternalRepositoryWrites: 0, OfficialMutationCount: 0, PromotionCount: 0}
	report.ReportDigest = sealedReportDigest(report)
	policy := PolicyEvidence{SourceAvailable: true, GeneratedAvailable: true,
		SourceDigest: digestValue("source"), GeneratedDigest: digestValue("generated")}
	envelope := Issue(subject, "run-1", 1, report.ReportDigest, policy)
	return Input{EnvelopeAvailable: true, ReportAvailable: true, Envelope: envelope,
		Report: report, Policy: policy,
		Invocation: Invocation{SubjectSHA: subject, RunID: "run-1", RunAttempt: 1}}
}

func TestBootstrapLowersResolution(t *testing.T) {
	receipt := Evaluate(fixtureInput())
	if receipt.Decision != DecisionFailClosed || receipt.Resolution != ResolutionUnknown ||
		receipt.Completed != 9 || receipt.OpenClaims != 1 {
		t.Fatalf("unexpected bootstrap receipt: %#v", receipt)
	}
	if len(receipt.Unknowns) != 1 || receipt.Unknowns[0].Stage != "AUTHORIZE/policy-foundation" {
		t.Fatalf("unknown provenance was not retained: %#v", receipt.Unknowns)
	}
}
