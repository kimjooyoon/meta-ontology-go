package externalconformanceactivation

func validMerge(proof mergeProof) bool {
	return proof.Schema == "gooo/github-pull-request-merge-proof/v1" &&
		proof.Repository == "kimjooyoon/meta-ontology-go" && proof.PullRequest == 474 &&
		proof.State == "MERGED" && proof.BaseSHA == EligibilityAssuranceSHA &&
		proof.HeadSHA == EligibilitySubjectSHA && proof.MergeCommitSHA == PredecessorSHA &&
		proof.MergedAt == "2026-08-24T18:42:01Z"
}
