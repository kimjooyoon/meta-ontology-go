package selfimprovementtransport

type Report struct {
	Schema                  string            `json:"schema"`
	MetricID                string            `json:"metric_id"`
	Contract                ContractEvidence  `json:"contract"`
	SubjectSHA              string            `json:"subject_sha"`
	OrchestrationHeadSHA    string            `json:"orchestration_head_sha"`
	SourceObservationDigest string            `json:"source_observation_digest"`
	ActualArchiveDigest     string            `json:"actual_archive_digest"`
	Decision                string            `json:"decision"`
	Resolution              string            `json:"resolution"`
	Reason                  string            `json:"reason"`
	Coordinate              Coordinate        `json:"coordinate"`
	Producer                ProducerReceipt   `json:"producer"`
	Transport               TransportMetadata `json:"transport"`
	Obligations             []Obligation      `json:"obligations"`
	OpenObligationIDs       []string          `json:"open_obligation_ids"`
	Metrics                 Metrics           `json:"metrics"`
	NotClaimed              []string          `json:"not_claimed"`
	Digest                  string            `json:"digest"`
}

type observationHeader struct {
	Schema     string `json:"schema"`
	SubjectSHA string `json:"subject_sha"`
}
