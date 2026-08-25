package languageprofileexperiment

import (
	"encoding/hex"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
)

func Evaluate(input Input) Report {
	if reason := inputReason(input); reason != "" {
		return closed(input, "EXACT", reason, 0)
	}
	if topDecisionUnknown(input) {
		return closed(input, "LOWER_RESOLUTION", "PROFILE_DECISION_UNKNOWN", 1)
	}
	if languageprofile.Validate(input.First) != nil || languageprofile.Validate(input.Replay) != nil ||
		languageprofile.Validate(input.UnknownEntry) != nil {
		return closed(input, "EXACT", "PROFILE_RECEIPT_INVALID", 0)
	}
	facts := observeFacts(input)
	indicators := buildIndicators(input, facts)
	summary := summarize(input, facts, indicators)
	report := Report{
		Schema: ReportSchema, Decision: "PASS", Resolution: "EXACT", Reason: "LANGUAGE_PROFILE_EXPERIMENT_OBSERVED",
		Interpretation: "RUNNER_SCOPED_PROFILE_OBSERVED", SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		ResourceObservationMode: "RUNNER_SCOPED_NONDETERMINISTIC", Summary: summary,
		Indicators: indicators, Views: buildViews(indicators), NotClaimed: languageprofile.DefaultNonClaims(),
		RepositoryWrites: summary.Effects.RepositoryWrites, MutationAuthority: summary.Effects.MutationAuthority,
		FactsDigest: digestValue(input),
	}
	report.Proofs = buildProofs(summary, report.FactsDigest)
	if summary.Coordinates.Satisfied != ExpectedIndicators {
		report.Decision, report.Reason, report.Interpretation = "FAIL_CLOSED", "LANGUAGE_PROFILE_CONTRACT_NOT_SATISFIED", "NO_LANGUAGE_QUALITY_CLAIM"
	}
	return sealReport(report)
}

func inputReason(input Input) string {
	if reason := contractReason(input.Contract); reason != "" {
		return reason
	}
	if len(input.SubjectSHA) != 40 {
		return "PROFILE_SUBJECT_SHA_INVALID"
	}
	if _, err := hex.DecodeString(input.SubjectSHA); err != nil || !validDigest(input.ExecutableDigest) {
		return "PROFILE_SUBJECT_IDENTITY_INVALID"
	}
	return ""
}

func topDecisionUnknown(input Input) bool {
	for _, decision := range []string{input.First.Decision, input.Replay.Decision, input.UnknownEntry.Decision} {
		if decision != "PASS" && decision != "FAIL_CLOSED" {
			return true
		}
	}
	return false
}

func closed(input Input, resolution, reason string, unknowns int) Report {
	facts := observeFacts(input)
	summary := Summary{Coordinates: Counter{Total: ExpectedIndicators}, Unknowns: unknowns,
		Effects: EffectSummary{RepositoryWrites: facts.writes, MutationAuthority: facts.mutation},
		NotClaimed: ExpectedNonClaims, Compiler: CompilerSummary{ExecutableDigest: input.ExecutableDigest}}
	report := Report{Schema: ReportSchema, Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Interpretation: "NO_LANGUAGE_QUALITY_CLAIM", SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		ResourceObservationMode: "RUNNER_SCOPED_NONDETERMINISTIC", Summary: summary,
		Indicators: []Indicator{}, Views: []View{}, Proofs: []Proof{}, NotClaimed: languageprofile.DefaultNonClaims(),
		RepositoryWrites: facts.writes, MutationAuthority: facts.mutation, FactsDigest: digestValue(input)}
	return sealReport(report)
}

func buildProofs(summary Summary, evidence string) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "produce-runner-profile-receipts", EvidenceDigest: evidence,
			Passed: summary.Profiles == 2 && summary.Samples == 10 && strings.HasPrefix(summary.Compiler.ExecutableDigest, "sha256:")},
		{Choice: "COHERENCE", MetaOperation: "compare-profiled-execution-digests", EvidenceDigest: evidence,
			Passed: summary.SourceCoherence == 1 && summary.ExecutionDigestVariants == 1},
		{Choice: "REGRESSION", MetaOperation: "reject-unknown-profile-entry", EvidenceDigest: evidence,
			Passed: summary.UnknownEntryRejections == 1 && summary.Effects.RepositoryWrites == 0 && !summary.Effects.MutationAuthority},
	}
}
