package replay

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func executeInvalid(repository fs.FS, path, expected string) Result {
	result := Result{ObservedDecision: "DIAGNOSTIC_MISMATCH"}
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		result.ObservedDecision, result.Diagnostics = DecisionUnknown, []string{"io.read"}
		return result
	}
	result.SourceLines, result.SourceDigest = sourceLines(raw), digestBytes(raw)
	_, diagnostics := syntax.ParseFile(path, string(raw))
	for _, diagnostic := range diagnostics {
		code := string(diagnostic.Code)
		result.Diagnostics = append(result.Diagnostics, code)
		if code == expected && diagnostic.Severity == syntax.SeverityError {
			result.DiagnosticRejected = true
		}
	}
	if !diagnostics.HasErrors() {
		result.ObservedDecision = DecisionPass
	} else if result.DiagnosticRejected {
		result.ObservedDecision = DecisionClosed
	}
	return result
}
