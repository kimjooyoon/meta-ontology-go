package valueexecution

func (execution Execution) BindRuntimePlanDigest(digest string) (Execution, error) {
	if digest == "" {
		return execution, nil
	}
	if !validDigest(digest) {
		return execution, failAt(ReasonPlanInvalid, "PLAN", "bind-runtime-plan-digest", "runtime plan digest is invalid")
	}
	execution.RuntimePlanDigest = digest
	execution.ExecutionDigest = executionDigest(execution)
	return execution, nil
}

func (trace Continuation) BindRuntimePlanDigest(digest string) (Continuation, error) {
	if digest == "" {
		return trace, nil
	}
	if !validDigest(digest) {
		return trace, failAt(ReasonPlanInvalid, "PLAN", "bind-runtime-plan-digest", "runtime plan digest is invalid")
	}
	for index := range trace.Executions {
		bound, err := trace.Executions[index].BindRuntimePlanDigest(digest)
		if err != nil {
			return trace, err
		}
		trace.Executions[index] = bound
	}
	trace.RuntimePlanDigest = digest
	trace.Digest = continuationDigest(trace)
	return trace, nil
}

func continuationDigest(trace Continuation) string {
	trace.Digest = ""
	return digestValue(trace)
}
