package selfimprovementattestation

type ResolutionReceipt struct {
	Schema                  string            `json:"schema"`
	Metaprogram             string            `json:"metaprogram"`
	MetricID                string            `json:"metric_id"`
	Contract                Contract          `json:"contract"`
	SubjectSHA              string            `json:"subject_sha"`
	PriorReceiptDigest      string            `json:"prior_receipt_digest"`
	SourceArchiveDigest     string            `json:"source_archive_digest"`
	Decision                string            `json:"decision"`
	Resolution              string            `json:"resolution"`
	Reason                  string            `json:"reason"`
	Coordinate              Coordinate        `json:"coordinate"`
	Checker                 Checker           `json:"checker"`
	ProducerIdentity        ProducerIdentity  `json:"producer_identity"`
	Obligations             []Obligation      `json:"obligations"`
	ClaimTransitions        []ClaimTransition `json:"claim_transitions"`
	OpenObligationIDs       []string          `json:"open_obligation_ids"`
	PriorMetrics            Metrics           `json:"prior_metrics"`
	Metrics                 Metrics           `json:"metrics"`
	Views                   []ReaderView      `json:"views"`
	Proofs                  []Proof           `json:"proofs"`
	Authority               Authority         `json:"authority"`
	NotClaimed              []string          `json:"not_claimed"`
	Digest                  string            `json:"digest"`
}
