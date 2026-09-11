package selfimprovementloop

type CounterexampleResult struct {
	Checked  bool   `json:"checked"`
	Found    bool   `json:"found"`
	Evidence string `json:"evidence"`
}

type CIResult struct {
	Executed bool   `json:"executed"`
	Passed   bool   `json:"passed"`
	Evidence string `json:"evidence"`
}

type ReceiptInput struct {
	Captured bool   `json:"captured"`
	Digest   string `json:"digest"`
}

type ExactPair struct {
	Before []MetricSample `json:"before"`
	After  []MetricSample `json:"after"`
}

type MetricSample struct {
	Scenario        string `json:"scenario"`
	SourceDigest    string `json:"source_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	Metric          string `json:"metric"`
	Value           int64  `json:"value"`
}

type HumanDecision struct {
	Decision string `json:"decision"`
}
