package predecessorresolution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func validateSelectionReport(report predecessorselection.Report) error {
	if report.Schema != predecessorselection.Schema || report.ReportDigest == "" {
		return fmt.Errorf("selection contract malformed")
	}
	digest := report.ReportDigest
	report.ReportDigest = ""
	if digestJSON(report) != digest {
		return fmt.Errorf("selection digest mismatch")
	}
	if report.Decision == predecessorselection.DecisionSelected {
		if report.Reason != predecessorselection.ReasonSelected || report.Selected == nil {
			return fmt.Errorf("selected decision malformed")
		}
		return nil
	}
	if report.Decision != predecessorselection.DecisionFailClosed ||
		report.Selected != nil || !knownFailureReason(report.Reason) {
		return fmt.Errorf("fail-closed decision malformed")
	}
	return nil
}

func knownFailureReason(reason string) bool {
	switch reason {
	case predecessorselection.ReasonNotFound,
		predecessorselection.ReasonUnbound,
		predecessorselection.ReasonFailed,
		predecessorselection.ReasonProducer,
		predecessorselection.ReasonExpired,
		predecessorselection.ReasonInvalid,
		predecessorselection.ReasonAmbiguous,
		predecessorselection.ReasonWriteEffect,
		predecessorselection.ReasonHTTPFailure,
		predecessorselection.ReasonPaginationIncomplete,
		predecessorselection.ReasonPaginationRepeated,
		predecessorselection.ReasonPaginationCap,
		predecessorselection.ReasonPaginationMalformed,
		predecessorselection.ReasonPaginationOrigin,
		predecessorselection.ReasonPaginationRedirect,
		predecessorselection.ReasonPaginationDuplicate,
		predecessorselection.ReasonResponseMalformed,
		"ARTIFACT_HTTP_FAILURE", "ARTIFACT_NEXT_LINK_REPEATED",
		"ARTIFACT_PAGE_CAP_EXCEEDED", "ARTIFACT_LINK_MALFORMED",
		"ARTIFACT_LINK_ORIGIN_MISMATCH", "ARTIFACT_REDIRECT_ORIGIN_MISMATCH", "ARTIFACT_DUPLICATE_ID",
		"ARTIFACT_RESPONSE_MALFORMED", "ARTIFACT_PAGINATION_INCOMPLETE",
		"JOB_HTTP_FAILURE", "JOB_NEXT_LINK_REPEATED", "JOB_PAGE_CAP_EXCEEDED",
		"JOB_LINK_MALFORMED", "JOB_LINK_ORIGIN_MISMATCH", "JOB_REDIRECT_ORIGIN_MISMATCH", "JOB_DUPLICATE_ID",
		"JOB_RESPONSE_MALFORMED", "JOB_PAGINATION_INCOMPLETE",
		predecessorselection.ReasonArtifactPayload:
		return true
	default:
		return false
	}
}
