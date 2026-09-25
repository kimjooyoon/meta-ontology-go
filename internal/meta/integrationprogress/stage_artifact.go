package integrationprogress

import "time"

func evaluateArtifact(value PullObservation, spec StageSpec, upstream Cell, timing runTiming) (Cell, time.Time) {
	if upstream.State != StateClosed {
		return dependentCell(value.Number, spec, upstream), time.Time{}
	}
	run := value.AuthoritativeRun
	if run.ArtifactQueryFailure != "" {
		return cell(value.Number, spec, StateUnknown, run.ArtifactQueryFailure, ""), time.Time{}
	}
	if run.ArtifactsTotal != run.ArtifactsConsumed {
		return cell(value.Number, spec, StateUnknown, "ARTIFACT_PAGINATION_INCOMPLETE", ""), time.Time{}
	}
	if run.ArtifactMatches == 0 && run.Artifact == nil {
		return cell(value.Number, spec, StateOpen, "EVIDENCE_ARTIFACT_NOT_FOUND", ""), time.Time{}
	}
	if run.ArtifactMatches != 1 || run.Artifact == nil {
		return cell(value.Number, spec, StateRefuted, "EVIDENCE_ARTIFACT_AMBIGUOUS", ""), time.Time{}
	}
	artifact := run.Artifact
	if artifact.Expired {
		return cell(value.Number, spec, StateOpen, "EVIDENCE_ARTIFACT_EXPIRED", ""), time.Time{}
	}
	if artifact.ID < 1 || artifact.Name != ArtifactPrefix+value.HeadSHA {
		return cell(value.Number, spec, StateRefuted, "ARTIFACT_IDENTITY_CONTRADICTION", ""), time.Time{}
	}
	if artifact.HeadSHA == "" {
		return cell(value.Number, spec, StateUnknown, "ARTIFACT_HEAD_UNKNOWN", ""), time.Time{}
	}
	created, ok := parseTime(artifact.CreatedAt)
	if !ok || artifact.HeadSHA != value.HeadSHA || created.Before(timing.Created) {
		return cell(value.Number, spec, StateRefuted, "ARTIFACT_SUBJECT_CONTRADICTION", ""), time.Time{}
	}
	return cell(value.Number, spec, StateClosed, "EVIDENCE_ARTIFACT_REACHABLE", "artifact/"+i64toa(artifact.ID)), created
}
