package languageresourcebudgetconsumer

type Binding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type SourceMeaning struct {
	Package        string    `json:"package"`
	Namespace      string    `json:"namespace"`
	Activity       string    `json:"activity"`
	Inputs         []Binding `json:"inputs"`
	Output         Binding   `json:"output"`
	SourceDigest   string    `json:"source_digest"`
	SemanticDigest string    `json:"semantic_digest"`
}

type ArtifactMeaning struct {
	Activity string    `json:"activity"`
	Inputs   []Binding `json:"inputs"`
	Output   Binding   `json:"output"`
	Decision string    `json:"decision"`
	Reason   string    `json:"reason"`
}

type ResourceOperation struct {
	Operation         string `json:"operation"`
	Samples           int    `json:"samples"`
	WallMaxNS         int64  `json:"wall_max_ns"`
	PeakRSSMaxKiB     int64  `json:"peak_rss_max_kib"`
	ReceiptMaxBytes   int64  `json:"receipt_max_bytes"`
	GeneratedMaxBytes int64  `json:"generated_max_bytes"`
	BudgetViolations  int    `json:"budget_violations"`
}

type ResourceEnvelope struct {
	Decision        string              `json:"decision"`
	Resolution      string              `json:"resolution"`
	Operations      int                 `json:"operations"`
	Samples         int                 `json:"samples"`
	ExpectedSamples int                 `json:"expected_samples"`
	Runner          Runner              `json:"runner"`
	PerOperation    []ResourceOperation `json:"per_operation"`
}

type ImportBoundary struct {
	ProducerImplementationImported bool `json:"producer_implementation_imported"`
	ReducerImplementationImported  bool `json:"reducer_implementation_imported"`
	Independent                    bool `json:"independent"`
	Numerator                      int  `json:"numerator"`
	Denominator                    int  `json:"denominator"`
}

type Provenance struct {
	RawSourceFiles      int    `json:"raw_source_files"`
	RawOperationOutputs int    `json:"raw_operation_outputs"`
	RawResourceSamples  int    `json:"raw_resource_samples"`
	EvidenceDigest      string `json:"evidence_digest"`
	ConsumerPackage     string `json:"consumer_package"`
}

type ClaimTransition struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	Evidence       string `json:"evidence"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}

type Report struct {
	Schema             string              `json:"schema"`
	Label              string              `json:"label"`
	EvidenceClass      string              `json:"evidence_class"`
	Decision           string              `json:"decision"`
	Resolution         string              `json:"resolution"`
	Reason             string              `json:"reason"`
	SemanticDecision   string              `json:"semantic_decision"`
	SemanticResolution string              `json:"semantic_resolution"`
	SemanticReason     string              `json:"semantic_reason"`
	Source             SourceMeaning       `json:"source"`
	Artifact           ArtifactMeaning     `json:"artifact"`
	Resource           ResourceEnvelope    `json:"resource"`
	WriteSet           WriteSetObservation `json:"write_set"`
	Imports            ImportBoundary      `json:"imports"`
	Provenance         Provenance          `json:"provenance"`
	ClaimTransitions   []ClaimTransition   `json:"claim_transitions"`
	FactsDigest        string              `json:"facts_digest"`
	Digest             string              `json:"digest"`
}
