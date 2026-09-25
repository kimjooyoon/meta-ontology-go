package adapter

import (
	"strings"
	"testing"
)

func newClosedReceipt(t *testing.T) (ProvenanceReceipt, Request, NoWriteObservation) {
	return newRejectedReceipt(t, RejectionClosed, ReceiptOutcomeClosed)
}
func newRejectedReceipt(t *testing.T, reason RejectionKind, outcome ReceiptOutcome) (ProvenanceReceipt, Request, NoWriteObservation) {
	t.Helper()
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	workflow := verifiedTestWorkflow(request)
	head := workflow.HeadSHA
	observation, err := observer.CaptureRejected(reason)
	if err != nil {
		t.Fatal(err)
	}
	digests, err := observation.StateDigests(request)
	if err != nil {
		t.Fatal(err)
	}
	return ProvenanceReceipt{
		Schema: ProvenanceReceiptSchema, Repository: "caller/repository",
		BaseSHA: strings.Repeat("a", 40), HeadSHA: head,
		EventRef: "event-002", CheckoutRef: "refs/pull/104/merge",
		Run: workflow.Run, Attempt: 1, ArtifactCount: workflow.ArtifactCount,
		Jobs:               successfulReceiptJobs(head),
		Binding:            requestObservationBinding(request),
		PreconditionDigest: digestBytes([]byte("precondition")),
		BeforeStateDigest:  digests.Before, AfterStateDigest: digests.After,
		Outcome: outcome, WriteEffect: ReceiptWriteEffectNone,
		Predecessors: []ReceiptPredecessor{}, ProvenanceStatus: ReceiptProvenanceVerified,
	}, request, observation
}
func successfulReceiptJobs(head string) []ReceiptJob {
	jobs := make([]ReceiptJob, 0, len(receiptJobNames))
	for _, name := range receiptJobNames {
		jobs = append(jobs, ReceiptJob{Name: name, HeadSHA: head, Status: "completed", Conclusion: "success"})
	}
	return jobs
}
