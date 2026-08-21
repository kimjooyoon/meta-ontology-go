package formatter

import (
	"strings"
)

// HasErrors reports whether formatting should be considered unsuccessful.
func (d Diagnostics) HasErrors() bool {
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Error joins error diagnostics while ignoring warnings.
func (d Diagnostics) Error() error {
	if !d.HasErrors() {
		return nil
	}
	lines := make([]string, 0, len(d))
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			lines = append(lines, diagnostic.String())
		}
	}
	return diagnosticError(strings.Join(lines, "\n"))
}

type diagnosticError string

func (e diagnosticError) Error() string { return string(e) }
