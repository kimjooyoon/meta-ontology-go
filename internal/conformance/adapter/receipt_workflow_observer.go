package adapter

import "reflect"

func validateObservedWorkflow(receipt ProvenanceReceipt, workflow WorkflowBinding) error {
	switch workflow.Status {
	case WorkflowEvidenceMissing:
		return oracleError(OracleNW001, "observer workflow evidence is missing")
	case WorkflowEvidenceUnverified:
		return oracleError(OracleNW003, "observer workflow evidence is unverified")
	case WorkflowEvidenceVerified:
		if err := workflow.validate(); err != nil {
			return oracleError(OracleNW003, "observer workflow evidence: "+err.Error())
		}
		if receipt.ProvenanceStatus != ReceiptProvenanceVerified {
			return oracleError(OracleNW003, "receipt provenance is unverified")
		}
		if !workflowMatchesReceipt(workflow, receipt) {
			return oracleError(OracleNW002, "receipt workflow binding is stale")
		}
		return nil
	default:
		return oracleError(OracleNW003, "observer workflow evidence status is invalid")
	}
}

func workflowMatchesReceipt(workflow WorkflowBinding, receipt ProvenanceReceipt) bool {
	return workflow.Repository == receipt.Repository &&
		workflow.BaseSHA == receipt.BaseSHA && workflow.HeadSHA == receipt.HeadSHA &&
		workflow.EventRef == receipt.EventRef && workflow.CheckoutRef == receipt.CheckoutRef &&
		workflow.Run == receipt.Run && workflow.Attempt == receipt.Attempt &&
		workflow.ArtifactCount == receipt.ArtifactCount &&
		reflect.DeepEqual(workflow.Jobs, receipt.Jobs)
}
