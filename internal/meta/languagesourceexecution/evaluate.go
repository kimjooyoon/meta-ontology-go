package languagesourceexecution

func Evaluate(input Input) Artifact {
	if _, err := DecodeContract(mustJSON(input.Contract)); err != nil {
		return failedContract(input.HeadSHA, err.Error())
	}
	cases, summary := evaluateCases(input)
	decision, resolution, reason := "PASS", "EXACT", "SOURCE_EXECUTION_CONTRACT_SATISFIED"
	if summary.Unknowns != 0 {
		decision, resolution, reason = "FAIL_CLOSED", "LOWER_RESOLUTION", "SOURCE_EXECUTION_EVIDENCE_UNKNOWN"
	} else if summary.NotSatisfied != 0 || summary.RepositoryWrites != 0 || summary.MutationAuthorities != 0 {
		decision, resolution, reason = "FAIL_CLOSED", "INVARIANT_ONLY", "SOURCE_EXECUTION_CONTRACT_VIOLATED"
	}
	evidence := digestValue(struct {
		Head    string
		Cases   []CaseResult
		Summary Summary
	}{input.HeadSHA, cases, summary})
	artifact := Artifact{
		Schema: ArtifactSchema, HeadSHA: input.HeadSHA, Decision: decision,
		Resolution: resolution, Reason: reason, ContractDigest: digestValue(input.Contract),
		Cases: cases, Summary: summary, Indicators: indicators(summary), Proofs: proofs(summary, evidence),
		RepositoryWrites: summary.RepositoryWrites, MutationAuthority: summary.MutationAuthorities != 0,
		NotClaimed: []string{"value-level computation", "external dependency execution", "cross-run resource improvement"},
	}
	artifact.Digest = artifactDigest(artifact)
	return artifact
}

func failedContract(head, reason string) Artifact {
	artifact := Artifact{Schema: ArtifactSchema, HeadSHA: head, Decision: "FAIL_CLOSED",
		Resolution: "LOWER_RESOLUTION", Reason: reason, Summary: Summary{CasesTotal: 4, Unknowns: 4},
		Cases: []CaseResult{}, Indicators: []Indicator{}, Proofs: []Proof{},
		NotClaimed: []string{"runtime capability without a canonical contract"}}
	artifact.Digest = artifactDigest(artifact)
	return artifact
}
