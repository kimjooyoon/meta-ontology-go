package adapter

// CaptureUnverifiedWorkflow records advisory workflow data without promotion.
func (o *NoWriteObserver) CaptureUnverifiedWorkflow(workflow WorkflowBinding) error {
	if workflow.Status == WorkflowEvidenceVerified {
		return oracleError(OracleNW003, "public workflow capture cannot verify evidence")
	}
	return o.captureWorkflow(workflow)
}

func (o *NoWriteObserver) captureVerifiedWorkflow(evidence verifiedWorkflowEvidence) error {
	return o.captureWorkflow(evidence.binding)
}

func (o *NoWriteObserver) captureWorkflow(workflow WorkflowBinding) error {
	if o == nil || o.stamp == nil || o.finished {
		return oracleError(OracleNW003, "observer workflow capture is closed")
	}
	if o.workflowCaptured {
		return oracleError(OracleNW003, "observer workflow capture is immutable")
	}
	if err := workflow.validate(); err != nil {
		return oracleError(OracleNW003, "workflow evidence: "+err.Error())
	}
	o.workflow = workflow.clone()
	o.workflowCaptured = true
	return nil
}
