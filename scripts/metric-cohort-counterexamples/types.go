package main

import (
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

type origin struct {
	Repository     string `json:"repository"`
	SubjectSHA     string `json:"subject_sha"`
	RunID          int64  `json:"run_id"`
	Attempt        int    `json:"run_attempt"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
}

type fixture struct {
	Receipt  co.Receipt
	Raw      []byte
	Inputs   co.SourceInputs
	Concept  []byte
	Contract []metricstrategy.ConceptOperationSpec
}

type variant struct {
	Receipt           []byte
	Inputs            co.SourceInputs
	Producer          *co.Receipt
	ProducerError     string
	ProducerAttempted bool
}

type caseResult struct {
	ID                      string            `json:"id"`
	ElapsedNanoseconds      int64             `json:"elapsed_nanoseconds"`
	Kind                    string            `json:"kind"`
	ExpectedReason          string            `json:"expected_reason"`
	ObservedReason          string            `json:"observed_reason"`
	ProducerError           string            `json:"producer_error,omitempty"`
	ExpectedProducerError   string            `json:"expected_producer_error,omitempty"`
	ProducerAttempted       bool              `json:"producer_attempted"`
	ConsumerAttempted       bool              `json:"consumer_attempted"`
	ProducerReceipt         *co.Receipt       `json:"producer_receipt,omitempty"`
	ClaimedExpectedCount    *int              `json:"claimed_expected_count,omitempty"`
	ClaimedCoverageBPS      *int              `json:"claimed_coverage_bps,omitempty"`
	EnvelopeAccepted        bool              `json:"envelope_accepted"`
	ReadinessCredit         bool              `json:"readiness_credit"`
	ReadinessResultProduced bool              `json:"readiness_result_produced"`
	InputDigests            map[string]string `json:"input_digests"`
	HarnessError            string            `json:"harness_error,omitempty"`
	Passed                  bool              `json:"passed"`
}
