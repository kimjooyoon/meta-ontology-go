package languagedelivery

type ReadinessArtifact struct {
	Schema   string `json:"schema"`
	Decision string `json:"decision"`
	HeadSHA  string `json:"head_sha"`
	Report   struct {
		Schema           string                `json:"schema"`
		Decision         string                `json:"decision"`
		Obligations      []ReadinessObligation `json:"obligations"`
		RepositoryWrites int                   `json:"repository_writes"`
	} `json:"report"`
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
	Readiness   ReadinessArtifact
}
