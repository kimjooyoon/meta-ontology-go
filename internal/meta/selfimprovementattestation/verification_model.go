package selfimprovementattestation

type VerificationItem struct {
	VerificationResult VerificationResult `json:"verificationResult"`
}

type VerificationResult struct {
	Signature          Signature           `json:"signature"`
	Statement          Statement           `json:"statement"`
	VerifiedTimestamps []VerifiedTimestamp `json:"verifiedTimestamps"`
}

type Signature struct {
	Certificate Certificate `json:"certificate"`
}

type Certificate struct {
	SubjectAlternativeName   string `json:"subjectAlternativeName"`
	Issuer                   string `json:"issuer"`
	GitHubWorkflowTrigger    string `json:"githubWorkflowTrigger"`
	GitHubWorkflowSHA        string `json:"githubWorkflowSHA"`
	GitHubWorkflowName       string `json:"githubWorkflowName"`
	GitHubWorkflowRepository string `json:"githubWorkflowRepository"`
	GitHubWorkflowRef        string `json:"githubWorkflowRef"`
	BuildSignerURI           string `json:"buildSignerURI"`
	BuildSignerDigest        string `json:"buildSignerDigest"`
	RunnerEnvironment        string `json:"runnerEnvironment"`
	SourceRepositoryURI      string `json:"sourceRepositoryURI"`
	SourceRepositoryDigest   string `json:"sourceRepositoryDigest"`
	SourceRepositoryRef      string `json:"sourceRepositoryRef"`
	BuildConfigURI           string `json:"buildConfigURI"`
	BuildConfigDigest        string `json:"buildConfigDigest"`
	RunInvocationURI         string `json:"runInvocationURI"`
}

type Statement struct {
	PredicateType string    `json:"predicateType"`
	Subject       []Subject `json:"subject"`
}

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

type VerifiedTimestamp struct {
	Type      string `json:"type"`
	URI       string `json:"uri"`
	Timestamp string `json:"timestamp"`
}
