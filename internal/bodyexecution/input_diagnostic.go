package bodyexecution

import (
	"errors"
	"fmt"
)

// InputDiagnostic describes one source-facing input refusal. It is an output
// projection: the code and keys do not participate in case validation.
type InputDiagnostic struct {
	Stage         string   `json:"stage"`
	Code          string   `json:"code"`
	CaseIndex     int      `json:"case_index"`
	Key           string   `json:"key,omitempty"`
	DeclaredKey   string   `json:"declared_key,omitempty"`
	AvailableKeys []string `json:"available_keys,omitempty"`
	Message       string   `json:"message"`
}

func (diagnostic *InputDiagnostic) Error() string { return diagnostic.Message }

func inputCaseDiagnostic(index int, err error) error {
	var diagnostic *InputDiagnostic
	if !errors.As(err, &diagnostic) {
		return fmt.Errorf("case %d: %w", index, err)
	}
	withCase := *diagnostic
	withCase.CaseIndex = index
	withCase.Message = fmt.Sprintf("case %d: %s", index, err)
	return &withCase
}
