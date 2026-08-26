package semantic

import (
	"strings"
)

func (e *InferencePathErrors) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrInferencePath.Error()
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, issue.Code+": "+issue.Detail)
	}
	return ErrInferencePath.Error() + ": " + strings.Join(parts, "; ")
}
func (e *InferencePathErrors) Unwrap() error { return ErrInferencePath }
func (e *InferencePathErrors) add(code string, record ID, detail string) {
	e.Issues = append(e.Issues, InferencePathIssue{Code: code, Record: record, Detail: detail})
}
