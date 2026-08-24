package sourceauthorityupstream

import (
	"fmt"
	"net/url"
)

func validate(policy Policy, request Request) string {
	if !validPolicy(policy) {
		return ReasonPolicyInvalid
	}
	if request.Schema != RequestSchema || !validHex(request.SubjectSHA, 40) {
		return ReasonRequestInvalid
	}
	if request.SourceRef != policy.SourceRef || request.AuthorityRef != policy.AuthorityRef {
		return ReasonAuthorityScopeMismatch
	}
	if request.URL != policy.URL || request.Authority != policy.Authority || request.Selection != policy.Selection {
		return ReasonAuthorityScopeMismatch
	}
	parsed, err := url.Parse(request.URL)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "raw.githubusercontent.com" {
		return ReasonAuthorityScopeMismatch
	}
	expectedPath := fmt.Sprintf("/%s/%s/%s", policy.Authority.Repository, policy.Authority.Revision, policy.Authority.Path)
	if parsed.Path != expectedPath || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return ReasonAuthorityScopeMismatch
	}
	return ""
}

func validPolicy(policy Policy) bool {
	if policy.SourceRef == "" || policy.AuthorityRef == "" || policy.URL == "" {
		return false
	}
	if policy.Authority.Repository == "" || policy.Authority.Path == "" || !validHex(policy.Authority.Revision, 40) {
		return false
	}
	if policy.Selection.StartLine < 1 || policy.Selection.EndLine < policy.Selection.StartLine {
		return false
	}
	return validDigest(policy.ExpectedDigest) && policy.ExpectedBytes > 0
}

func validDigest(value string) bool {
	const prefix = "sha256:"
	if len(value) != len(prefix)+64 || value[:len(prefix)] != prefix {
		return false
	}
	return validHex(value[len(prefix):], 64)
}

func validHex(value string, size int) bool {
	if len(value) != size {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	return true
}
