package integrationprogress

import "time"

func evaluateMerge(value PullObservation, spec StageSpec, upstream Cell) (Cell, time.Time) {
	if upstream.State != StateClosed {
		return dependentCell(value.Number, spec, upstream), time.Time{}
	}
	if value.MergedAt == "" {
		return cell(value.Number, spec, StateOpen, "PULL_NOT_MERGED", ""), time.Time{}
	}
	merged, ok := parseTime(value.MergedAt)
	if !ok || value.State != "closed" {
		return cell(value.Number, spec, StateRefuted, "MERGE_STATE_CONTRADICTION", ""), time.Time{}
	}
	return cell(value.Number, spec, StateClosed, "MERGE_REALIZED", pullRef(value.Number)), merged
}

func evaluateLink(number int, spec StageSpec, artifact, merge Cell, artifactAt, mergedAt time.Time) Cell {
	if merge.State == StateOpen {
		return cell(number, spec, StateOpen, "MERGE_NOT_REALIZED", "")
	}
	if merge.State != StateClosed {
		return dependentCell(number, spec, merge)
	}
	if artifact.State != StateClosed {
		return dependentCell(number, spec, artifact)
	}
	if mergedAt.Before(artifactAt) {
		return cell(number, spec, StateRefuted, "MERGED_BEFORE_EVIDENCE", "")
	}
	return cell(number, spec, StateClosed, "MERGED_EVIDENCE_LINKED", pullRef(number))
}
