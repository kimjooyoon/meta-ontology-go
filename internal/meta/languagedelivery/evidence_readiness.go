package languagedelivery

type ReadinessArtifact struct {
	Schema   string `json:"schema"`
	HeadSHA  string `json:"head_sha"`
	Snapshot struct {
		Schema           string                `json:"schema"`
		Decision         string                `json:"decision"`
		Obligations      []ReadinessObligation `json:"obligations"`
		RepositoryWrites int                   `json:"repository_writes"`
	} `json:"snapshot"`
}

type ReadinessObligation struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type decodedEvidence struct {
	Journey     JourneyReceipt
	Conformance ConformanceReceipt
	LSP         LSPReceipt
	Release     ReleaseReceipt
	Execution   ExecutionReceipt
	Profile     ProfileReceipt
	Debug       DebugReceipt
	Readiness   ReadinessArtifact
}
