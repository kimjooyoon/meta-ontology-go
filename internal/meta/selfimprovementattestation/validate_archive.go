package selfimprovementattestation

import (
	"fmt"
	"reflect"
)

func validateArchive(request Request) error {
	receipt := request.TransportReceipt
	if request.ArchiveDigest != receipt.ActualArchiveDigest {
		return fmt.Errorf("ATTESTED_ARCHIVE_DIGEST_MISMATCH")
	}
	if request.ArchiveDigest != receipt.Transport.ArtifactDigest {
		return fmt.Errorf("TRANSPORT_ARTIFACT_DIGEST_MISMATCH")
	}
	if !reflect.DeepEqual(request.ArchiveProducer, receipt.Producer) {
		return fmt.Errorf("SIGNED_PRODUCER_RECEIPT_MISMATCH")
	}
	if request.ArchiveObservationDigest != receipt.SourceObservationDigest {
		return fmt.Errorf("SIGNED_OBSERVATION_DIGEST_MISMATCH")
	}
	workflow := receipt.Transport.Repository + "/" + receipt.Transport.WorkflowPath + "@"
	if len(receipt.Producer.WorkflowRef) <= len(workflow) || receipt.Producer.WorkflowRef[:len(workflow)] != workflow {
		return fmt.Errorf("PRODUCER_WORKFLOW_REF_MISMATCH")
	}
	return nil
}
