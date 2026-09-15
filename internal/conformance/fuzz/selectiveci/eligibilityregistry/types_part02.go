package eligibilityregistry

type CoverageResult struct {
	RegistryDigest       Digest
	RegistrySourceDigest Digest
	CurrentSourceDigest  Digest
	Decision             Decision
	Reason               Reason
	Faults               []Reason
	FullSuiteRequired    bool
	ExecutionAuthorized  bool
	EnforcementEffect    EnforcementEffect
	RegisteredCount      uint64
	ObservedCount        uint64
	CoveredCount         uint64
	ResultDigest         Digest
	ReplayDigest         Digest
}
