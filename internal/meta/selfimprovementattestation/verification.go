package selfimprovementattestation

import (
	"fmt"
	"sort"
	"strings"
)

func selectVerification(request Request) (*VerificationResult, []string) {
	matches := make([]*VerificationResult, 0, 1)
	reasons := map[string]struct{}{}
	for index := range request.Verification {
		result := &request.Verification[index].VerificationResult
		mismatches := verificationMismatches(request, result)
		if len(mismatches) == 0 {
			matches = append(matches, result)
			continue
		}
		for _, reason := range mismatches {
			reasons[reason] = struct{}{}
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return nil, []string{"PRODUCER_ATTESTATION_AMBIGUOUS"}
	}
	ordered := make([]string, 0, len(reasons))
	for reason := range reasons {
		ordered = append(ordered, reason)
	}
	sort.Strings(ordered)
	if len(ordered) == 0 {
		ordered = append(ordered, "PRODUCER_ATTESTATION_NOT_FOUND")
	}
	return nil, ordered
}

func verificationMismatches(request Request, result *VerificationResult) []string {
	receipt, certificate := request.TransportReceipt, result.Signature.Certificate
	producer, transport := receipt.Producer, receipt.Transport
	ref := strings.SplitN(producer.WorkflowRef, "@", 2)
	if len(ref) != 2 {
		return []string{"PRODUCER_WORKFLOW_REF_INVALID"}
	}
	expectedURI := "https://github.com/" + producer.WorkflowRef
	expectedRun := fmt.Sprintf("https://github.com/%s/actions/runs/%d/attempts/%d", transport.Repository, producer.RunID, producer.RunAttempt)
	checks := []struct{ failed bool; reason string }{
		{certificate.Issuer != oidcIssuer, "PRODUCER_OIDC_ISSUER_MISMATCH"},
		{certificate.SubjectAlternativeName != expectedURI || certificate.BuildSignerURI != expectedURI, "PRODUCER_SIGNER_IDENTITY_MISMATCH"},
		{certificate.GitHubWorkflowRepository != transport.Repository, "PRODUCER_WORKFLOW_REPOSITORY_MISMATCH"},
		{certificate.GitHubWorkflowSHA != producer.WorkflowSHA || certificate.SourceRepositoryDigest != producer.WorkflowSHA, "PRODUCER_WORKFLOW_SHA_MISMATCH"},
		{certificate.BuildSignerDigest != producer.WorkflowSHA || certificate.BuildConfigDigest != producer.WorkflowSHA, "PRODUCER_BUILD_DIGEST_MISMATCH"},
		{certificate.GitHubWorkflowRef != ref[1] || certificate.SourceRepositoryRef != ref[1], "PRODUCER_WORKFLOW_REF_MISMATCH"},
		{certificate.SourceRepositoryURI != producer.RepositoryURI, "PRODUCER_SOURCE_REPOSITORY_MISMATCH"},
		{certificate.RunInvocationURI != expectedRun, "PRODUCER_RUN_INVOCATION_MISMATCH"},
		{certificate.RunnerEnvironment != "github-hosted", "PRODUCER_RUNNER_ENVIRONMENT_MISMATCH"},
		{certificate.GitHubWorkflowTrigger != "workflow_run", "PRODUCER_WORKFLOW_TRIGGER_MISMATCH"},
		{result.Statement.PredicateType != slsaPredicate, "PRODUCER_PREDICATE_TYPE_MISMATCH"},
		{len(result.VerifiedTimestamps) == 0, "PRODUCER_VERIFIED_TIMESTAMP_MISSING"},
		{!subjectMatches(result.Statement.Subject, producer.ArtifactName, request.ArchiveDigest), "ATTESTED_SUBJECT_DIGEST_MISMATCH"},
	}
	var reasons []string
	for _, check := range checks {
		if check.failed {
			reasons = append(reasons, check.reason)
		}
	}
	return reasons
}

func subjectMatches(subjects []Subject, name, digest string) bool {
	if len(subjects) != 1 || !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	return subjects[0].Name == name && subjects[0].Digest["sha256"] == strings.TrimPrefix(digest, "sha256:")
}
