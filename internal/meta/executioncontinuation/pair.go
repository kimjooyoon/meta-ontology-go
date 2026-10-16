package executioncontinuation

import (
	"io/fs"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
	input "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

type Pair struct {
	Grant                   grant.LiveReport                  `json:"grant"`
	Contract                contract.LiveReport               `json:"contract"`
	AuthorizationRequest    candidate.AuthorizationRequest    `json:"authorization_request"`
	AuthorizationResolution candidate.AuthorizationResolution `json:"authorization_resolution"`
}

// ValidatePair verifies source payloads, rather than trusting projected validity flags.
func ValidatePair(root fs.FS, p Pair, m Metadata) (*input.ExecutionInput, error) {
	if err := m.Validate(); err != nil {
		return nil, refuted(err.Error())
	}
	if err := ValidateGrant(root, p.Grant, m.Source); err != nil {
		return nil, err
	}
	if candidate.ValidateAuthorizationRequest(p.AuthorizationRequest) != nil || candidate.VerifyAuthorizationResolution(p.AuthorizationRequest, p.AuthorizationResolution) != nil {
		return nil, refuted("AUTHORIZATION_PAIR_INVALID")
	}
	if !reflect.DeepEqual(grant.ProjectV24(p.AuthorizationRequest, p.AuthorizationResolution), p.Grant.Request.V24) {
		return nil, refuted("AUTHORIZATION_ORIGIN_MISMATCH")
	}
	c := p.Contract.ContractResolution
	if contract.VerifyResolution(c) != nil || !p.Contract.Verification.Verified || p.Contract.Verification.ContractDigest != c.Digest || !reflect.DeepEqual(grant.ProjectV25(c), p.Grant.Request.V25) {
		return nil, refuted("CONTRACT_ORIGIN_MISMATCH")
	}
	source := p.Grant.Request.Source
	if source.Repository != m.Source.Repository || source.WorkflowRunID != m.ContractRunID || source.WorkflowRunAttempt != m.ContractAttempt || source.ArtifactID != m.ContractArtifactID || source.ArtifactDigest != m.ContractDigest || source.ObservedArtifactDigest != source.ArtifactDigest || !source.ArtifactRetrieved || !source.ArtifactExpiryKnown || source.ArtifactExpired || source.ArtifactRetrievalError != "" {
		return nil, refuted("CONTRACT_ARCHIVE_ORIGIN_MISMATCH")
	}
	if policy.BindSubjects(m.Source.SubjectSHA, p.Grant.Request.V24.SubjectSHA, c.SubjectSHA) != "SUBJECT_BOUND" {
		return nil, refuted("SUBJECT_REFUTED")
	}
	if c.ExecutionInput == nil || input.Validate(*c.ExecutionInput) != nil {
		return nil, refuted("EXECUTION_INPUT_INVALID")
	}
	return c.ExecutionInput, nil
}
