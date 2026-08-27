package integrationprogress

func evaluateRun(value PullObservation, spec StageSpec, upstream Cell) (Cell, runTiming) {
	if upstream.State != StateClosed {
		return dependentCell(value.Number, spec, upstream), runTiming{}
	}
	if value.RunQueryFailure != "" {
		return cell(value.Number, spec, StateUnknown, value.RunQueryFailure, ""), runTiming{}
	}
	if value.RunsTotal != value.RunsConsumed {
		return cell(value.Number, spec, StateUnknown, "RUN_PAGINATION_INCOMPLETE", ""), runTiming{}
	}
	if value.AuthoritativeRun == nil {
		return cell(value.Number, spec, StateOpen, "AUTHORITATIVE_RUN_NOT_FOUND", ""), runTiming{}
	}
	run := value.AuthoritativeRun
	if run.ID < 1 || run.Name != WorkflowName || run.HeadSHA != value.HeadSHA {
		return cell(value.Number, spec, StateRefuted, "RUN_IDENTITY_CONTRADICTION", ""), runTiming{}
	}
	if run.Status != "completed" {
		if run.Status == "queued" || run.Status == "in_progress" || run.Status == "pending" {
			return cell(value.Number, spec, StateOpen, "AUTHORITATIVE_RUN_NOT_TERMINAL", ""), runTiming{}
		}
		return cell(value.Number, spec, StateUnknown, "RUN_STATUS_UNKNOWN", ""), runTiming{}
	}
	created, createdOK := parseTime(run.CreatedAt)
	started, startedOK := parseTime(run.StartedAt)
	updated, updatedOK := parseTime(run.UpdatedAt)
	if !createdOK || !startedOK || !updatedOK || started.Before(created) || updated.Before(started) || run.Conclusion == "" {
		return cell(value.Number, spec, StateRefuted, "RUN_TIMELINE_CONTRADICTION", ""), runTiming{}
	}
	return cell(value.Number, spec, StateClosed, "AUTHORITATIVE_RUN_TERMINAL", "run/"+i64toa(run.ID)),
		runTiming{Created: created, Started: started, Updated: updated}
}
