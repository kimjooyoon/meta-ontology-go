package selfimprovementattestation

import "strings"

func fixtureVerification(producer Producer, archiveDigest string) VerificationResult {
	signer := "https://github.com/" + producer.WorkflowRef
	return VerificationResult{
		Signature: Signature{Certificate: Certificate{
			SubjectAlternativeName: signer, Issuer: oidcIssuer, GitHubWorkflowTrigger: "workflow_run",
			GitHubWorkflowSHA: producer.WorkflowSHA, GitHubWorkflowName: "Self-improvement language observation",
			GitHubWorkflowRepository: "owner/repo", GitHubWorkflowRef: "refs/heads/dev",
			BuildSignerURI: signer, BuildSignerDigest: producer.WorkflowSHA, RunnerEnvironment: "github-hosted",
			SourceRepositoryURI: producer.RepositoryURI, SourceRepositoryDigest: producer.WorkflowSHA,
			SourceRepositoryRef: "refs/heads/dev", BuildConfigURI: signer, BuildConfigDigest: producer.WorkflowSHA,
			RunInvocationURI: "https://github.com/owner/repo/actions/runs/12/attempts/1",
		}},
		Statement: Statement{PredicateType: slsaPredicate, Subject: []Subject{{
			Name: producer.ArtifactName, Digest: map[string]string{"sha256": strings.TrimPrefix(archiveDigest, "sha256:")},
		}}},
		VerifiedTimestamps: []VerifiedTimestamp{{Type: "Tlog", URI: "https://rekor.sigstore.dev", Timestamp: "2026-01-01T00:00:00Z"}},
	}
}
