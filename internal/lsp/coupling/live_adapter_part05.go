package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
	"strings"
)

func validateSourceLocation(request LiveRequest, location SourceLocation) *liveIssue {
	if err := validateIdentity(location.StableID, "source location"); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location stable ID is malformed."}
	}
	if err := validateIdentity(location.SourceMapID, "source map"); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location map ID is malformed."}
	}
	if err := validateDigest(location.SourceMapDigest); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location map digest is malformed."}
	}
	if location.SourceMapDigest != request.Query.SourceMapDigest {
		return &liveIssue{OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The source location map is stale."}
	}
	if location.URI == "" || strings.TrimSpace(location.URI) != location.URI {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location URI is malformed."}
	}
	if err := validateRange(location.Range); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location range is malformed."}
	}
	return nil
}
func liveQueryFailure(request LiveRequest, explanation couplingexplain.Explanation) Result {
	outcome := OutcomeUnknown
	if explanation.Status == couplingexplain.StatusFailClosed {
		outcome = OutcomeFailClosed
	}
	code, severity, message := liveReasonDiagnostic(explanation.NoLink)
	if outcome == OutcomeFailClosed {
		severity = DiagnosticError
	}
	return liveFailure(request, outcome, code, severity, message)
}
func liveReasonDiagnostic(noLink *couplingexplain.NoLink) (string, int, string) {
	if noLink == nil {
		return DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query did not return a link decision."
	}
	switch noLink.Reason {
	case couplingexplain.ReasonAmbiguous:
		return DiagnosticAmbiguous, DiagnosticWarning, "The coupling query is ambiguous."
	case couplingexplain.ReasonStale:
		return DiagnosticStaleSnapshot, DiagnosticWarning, "The coupling query is stale."
	case couplingexplain.ReasonUnregistered:
		return DiagnosticNoBinding, DiagnosticWarning, "The code surface is not registered in the coupling snapshot."
	case couplingexplain.ReasonMissing:
		return DiagnosticLiveMissingMaterial, DiagnosticWarning, "The coupling query lacks verified link material."
	case couplingexplain.ReasonNotVerified:
		return DiagnosticLiveNotVerified, DiagnosticError, "The coupling query link was not independently verified."
	default:
		return DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query returned an invalid link reason."
	}
}
