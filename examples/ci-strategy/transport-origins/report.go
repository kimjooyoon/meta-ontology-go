package transportorigins

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// FromSavedReport projects retained bytes. It performs no registry operation;
// the report's observations and generation contexts remain reported content.
func FromSavedReport(raw []byte) (R_OriginView, error) {
	report, frame, err := decodeSavedReport(raw)
	if err != nil {
		return R_OriginView{}, err
	}
	return projectSavedReport(report, frame), nil
}

func decodeSavedReport(raw []byte) (transport.Report, R_SavedFrame, error) {
	var report transport.Report
	limit := ReportByteLimit()
	if limit < 1 || uint64(limit) > uint64(^uint(0)>>1) {
		return report, R_SavedFrame{}, fmt.Errorf("report byte limit is not representable on this host")
	}
	if err := completeness.CheckJSON(raw, int(limit)); err != nil {
		return report, R_SavedFrame{}, err
	}
	projection, complete, err := projectReportJSON(raw, reflect.TypeFor[transport.Report]())
	if err != nil {
		return report, R_SavedFrame{}, err
	}
	if err := json.Unmarshal(projection, &report); err != nil {
		return report, R_SavedFrame{}, err
	}
	frame := R_SavedFrame{F_report_sha256: contentDigest(raw), F_report_bytes: int64(len(raw)),
		F_fields_complete: complete, F_after_present: report.After != nil, F_next_present: report.Next != nil,
		F_decision_generation_sha256:   contextDigest(report.GenerationOrigin),
		F_outcome_generation_sha256:    contextDigest(report.OutcomeGenerationOrigin),
		F_projection_generation_sha256: contextDigest(GoooGenerationOrigin)}
	return report, frame, nil
}

func projectSavedReport(report transport.Report, frame R_SavedFrame) R_OriginView {
	var after R_TransportObservation
	var next R_TransportStep
	if report.After != nil {
		after = R_TransportObservation(*report.After)
	}
	if report.Next != nil {
		next = R_TransportStep(*report.Next)
	}
	return Evaluate(R_TransportScope(report.Scope), R_TransportObservation(report.Before), R_TransportStep(report.Decision),
		R_WriteRecord{F_attempted: report.Write.Attempted, F_status: report.Write.Status, F_error: report.Write.Error},
		after, next, R_Outcome(report.Outcome), frame)
}

func contentDigest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }

func contextDigest(value string) string {
	if value == "" {
		return ""
	}
	return contentDigest([]byte(value))
}
