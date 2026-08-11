package semantic

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrGraphInvalid      = errors.New("invalid semantic graph")
	ErrIdentityConflict  = errors.New("semantic identity conflict")
	ErrNameCollision     = errors.New("semantic name collision")
	ErrNodeNotFound      = errors.New("semantic node not found")
	ErrFactNotFound      = errors.New("semantic fact not found")
	ErrCandidateNotFound = errors.New("semantic candidate not found")
)

// ValidationIssue is one deterministic graph invariant violation.
type ValidationIssue struct {
	Code    string
	Message string
	Subject ID
	Object  ID
}

// ValidationErrors retains all graph violations in stable traversal order.
type ValidationErrors struct {
	Issues []ValidationIssue
}

func (e *ValidationErrors) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrGraphInvalid.Error()
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		if issue.Code == "" {
			parts = append(parts, issue.Message)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", issue.Code, issue.Message))
	}
	return ErrGraphInvalid.Error() + ": " + strings.Join(parts, "; ")
}

func (e *ValidationErrors) Unwrap() error {
	return ErrGraphInvalid
}

func (e *ValidationErrors) add(code, message string, subject, object ID) {
	e.Issues = append(e.Issues, ValidationIssue{
		Code: code, Message: message, Subject: subject, Object: object,
	})
}
