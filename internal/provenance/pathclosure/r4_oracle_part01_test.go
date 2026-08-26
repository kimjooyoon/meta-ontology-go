package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
)

// literalR4Oracle is intentionally a closed table: it is not derived from
// EvaluateR4 and cannot learn a new production false positive by reuse.
func literalR4Oracle(name string) (pathclosure.Status, string) {
	return map[string]struct {
			status pathclosure.Status
			code   string
		}{
			"complete":                     {pathclosure.PASS, pathclosure.CodeR4ProofValid},
			"wrong subject":                {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"wrong object endpoint":        {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"wrong canonical record bytes": {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"wrong predecessor":            {pathclosure.FAIL_CLOSED, pathclosure.CodeInvalidPath},
			"duplicate receipt":            {pathclosure.FAIL_CLOSED, pathclosure.CodeConflictingReceipt},
			"conflicting receipt":          {pathclosure.FAIL_CLOSED, pathclosure.CodeConflictingReceipt},
			"missing record":               {pathclosure.UNKNOWN, pathclosure.CodeMissingRecord},
			"missing evidence binding":     {pathclosure.UNKNOWN, pathclosure.CodeMissingEvidence},
			"missing provider binding":     {pathclosure.UNKNOWN, pathclosure.CodeMissingProvider},
			"stale provider phase digest":  {pathclosure.UNKNOWN, pathclosure.CodePhaseMismatch},
			"producer only effect claim":   {pathclosure.UNKNOWN, pathclosure.CodeMissingObserver},
		}[name].status, map[string]string{
			"complete": pathclosure.CodeR4ProofValid, "wrong subject": pathclosure.CodeInvalidPath, "wrong object endpoint": pathclosure.CodeInvalidPath,
			"wrong canonical record bytes": pathclosure.CodeInvalidPath, "wrong predecessor": pathclosure.CodeInvalidPath, "duplicate receipt": pathclosure.CodeConflictingReceipt,
			"conflicting receipt": pathclosure.CodeConflictingReceipt, "missing record": pathclosure.CodeMissingRecord, "missing evidence binding": pathclosure.CodeMissingEvidence,
			"missing provider binding":    pathclosure.CodeMissingProvider,
			"stale provider phase digest": pathclosure.CodePhaseMismatch, "producer only effect claim": pathclosure.CodeMissingObserver,
		}[name]
}
