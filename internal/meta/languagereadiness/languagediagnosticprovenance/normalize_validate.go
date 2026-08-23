package languagediagnosticprovenance

import "github.com/kimjooyoon/meta-ontology-go/internal/formatter"

func validateObservation(observation Observation) *ProvenanceError {
	if observation.Origin != "GO" && observation.Origin != "FORMATTER" {
		return provenanceError("PROVENANCE_ORIGIN_UNKNOWN")
	}
	switch observation.Stage {
	case "PARSE", "TYPE", "FORMAT":
	default:
		return provenanceError("PROVENANCE_STAGE_UNKNOWN")
	}
	if observation.Severity != formatter.SeverityError &&
		observation.Severity != formatter.SeverityWarning {
		return provenanceError("PROVENANCE_SEVERITY_UNKNOWN")
	}
	if observation.Code == "" {
		return provenanceError("PROVENANCE_CODE_UNKNOWN")
	}
	switch observation.Hardness {
	case "HARD", "SOFT", "NOT_APPLICABLE":
	default:
		return provenanceError("PROVENANCE_HARDNESS_UNKNOWN")
	}
	if !validPosition(observation.Physical.Start) {
		return provenanceError("PHYSICAL_POSITION_UNKNOWN")
	}
	if !validSpan(observation.Physical) {
		return provenanceError("PHYSICAL_RANGE_INVALID")
	}
	if !validLogicalPosition(observation.Logical.Start) {
		return provenanceError("LOGICAL_POSITION_UNKNOWN")
	}
	if !validLogicalSpan(observation.Logical) {
		return provenanceError("LOGICAL_RANGE_INVALID")
	}
	return nil
}

func validPosition(position Position) bool {
	return position.Filename != "" && position.Offset >= 0 &&
		position.Line > 0 && position.Column > 0
}

func validSpan(span Span) bool {
	if !validPosition(span.Start) || !validPosition(span.End) ||
		span.Start.Filename != span.End.Filename {
		return false
	}
	if span.End.Offset <= span.Start.Offset || span.End.Line < span.Start.Line {
		return false
	}
	if span.End.Line == span.Start.Line && span.End.Column <= span.Start.Column {
		return false
	}
	return true
}

// validLogicalPosition accepts Go's explicit column-zero representation for a
// valid position whose logical column is unavailable after a //line remap.
func validLogicalPosition(position Position) bool {
	if position.Column < 0 {
		return false
	}
	if position.Column == 0 {
		position.Column = 1
	}
	return validPosition(position)
}

// validLogicalSpan never invents a missing logical column. Both ends must use
// the same lower resolution, while byte, line, and filename invariants remain
// identical to an exact span.
func validLogicalSpan(span Span) bool {
	unknownStart := span.Start.Column == 0
	unknownEnd := span.End.Column == 0
	if unknownStart != unknownEnd {
		return false
	}
	if !unknownStart {
		return validSpan(span)
	}
	return validLogicalPosition(span.Start) &&
		validLogicalPosition(span.End) &&
		span.Start.Filename == span.End.Filename &&
		span.End.Offset > span.Start.Offset &&
		span.End.Line >= span.Start.Line
}
