package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementCounterexampleObservationSchema = "gooo.self-improvement.counterexample-observation.v1"

type SelfImprovementCounterexampleStatus string

const SelfImprovementCounterexampleStatusRetained SelfImprovementCounterexampleStatus = "RETAINED"
const SelfImprovementCounterexampleStatusUnknown SelfImprovementCounterexampleStatus = "UNKNOWN"
const SelfImprovementCounterexampleStatusNone SelfImprovementCounterexampleStatus = "NONE"

type SelfImprovementCounterexampleObservation struct {
	Schema              string                              `json:"schema"`
	SourceURI           string                              `json:"source_uri"`
	InputDigest         string                              `json:"input_digest"`
	CandidateDigest     string                              `json:"candidate_digest"`
	BeforeReceiptDigest string                              `json:"before_receipt_digest"`
	AfterReceiptDigest  string                              `json:"after_receipt_digest"`
	ObservedReason      string                              `json:"observed_reason"`
	Status              SelfImprovementCounterexampleStatus `json:"status"`
	Reason              string                              `json:"reason"`
	NonAuthorizing      bool                                `json:"non_authorizing"`
	Digest              string                              `json:"digest"`
}

// ObserveSelfImprovementCounterexample retains replayable regression evidence
// instead of allowing a later improvement to erase its original boundary.
func ObserveSelfImprovementCounterexample(
	sourceURI string,
	inputDigest string,
	observation ExecutedSelfImprovementObservation,
	beforeReceipt ExecutionOriginReceipt,
	afterReceipt ExecutionOriginReceipt,
) SelfImprovementCounterexampleObservation {
	counterexample := SelfImprovementCounterexampleObservation{
		Schema:              SelfImprovementCounterexampleObservationSchema,
		SourceURI:           sourceURI,
		InputDigest:         inputDigest,
		CandidateDigest:     observation.CandidateDigest,
		BeforeReceiptDigest: beforeReceipt.ReceiptDigest,
		AfterReceiptDigest:  afterReceipt.ReceiptDigest,
		ObservedReason:      observation.Reason,
		Status:              SelfImprovementCounterexampleStatusUnknown,
		Reason:              "COUNTEREXAMPLE_UNKNOWN",
		NonAuthorizing:      true,
	}
	switch {
	case sourceURI == "" || inputDigest == "":
		counterexample.Reason = "COUNTEREXAMPLE_CONTEXT_MISSING"
	case !observation.NonAuthorizing || observation.Digest == "":
		counterexample.Reason = "COUNTEREXAMPLE_OBSERVATION_UNTRUSTED"
	case beforeReceipt.Status != ExecutionOriginStatusBound || afterReceipt.Status != ExecutionOriginStatusBound:
		counterexample.Reason = "COUNTEREXAMPLE_RECEIPT_UNBOUND"
	case beforeReceipt.Phase != ExecutionPhaseCompleted || afterReceipt.Phase != ExecutionPhaseCompleted:
		counterexample.Reason = "COUNTEREXAMPLE_EXECUTION_INCOMPLETE"
	case beforeReceipt.ReceiptDigest == "" || afterReceipt.ReceiptDigest == "" || beforeReceipt.ExecutionDigest == "" || afterReceipt.ExecutionDigest == "":
		counterexample.Reason = "COUNTEREXAMPLE_RECEIPT_DIGEST_MISSING"
	case observation.Status == SelfImprovementStatusRegressed:
		counterexample.Status = SelfImprovementCounterexampleStatusRetained
		counterexample.Reason = "REGRESSION_COUNTEREXAMPLE_RETAINED"
	case observation.Status == SelfImprovementStatusUnknown:
		counterexample.Reason = "UNKNOWN_COUNTEREXAMPLE_RETAINED"
	default:
		counterexample.Status = SelfImprovementCounterexampleStatusNone
		counterexample.Reason = "NO_COUNTEREXAMPLE"
	}
	counterexample.Digest = selfImprovementCounterexampleObservationDigest(counterexample)
	return counterexample
}

func selfImprovementCounterexampleObservationDigest(observation SelfImprovementCounterexampleObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
