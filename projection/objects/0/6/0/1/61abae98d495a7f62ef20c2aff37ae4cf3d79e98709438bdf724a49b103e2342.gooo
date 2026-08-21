package analyzer

// HostStage identifies which implementation is producing compiler evidence.
type HostStage string

const (
	// StageGoHosted is the implemented reference stage. Go hosts analysis and
	// .gooo remains the authoritative source view.
	StageGoHosted HostStage = "go-hosted"
	// StageGoooHosted is the future self-hosted stage. It is intentionally
	// deferred until an independent comparison contract is implemented.
	StageGoooHosted HostStage = "gooo-hosted"
)

// ContractStatus distinguishes an implemented contract from an explicitly
// deferred future stage. Deferred is never treated as a successful stage.
type ContractStatus string

const (
	ContractImplemented ContractStatus = "implemented"
	ContractDeferred    ContractStatus = "deferred"
)

// ContractRequirement names evidence a host stage must provide before it can
// be promoted.
type ContractRequirement string

const (
	RequirementStableIdentity  ContractRequirement = "stable-identity"
	RequirementDeltaEvidence   ContractRequirement = "delta-evidence"
	RequirementSourceSpans     ContractRequirement = "source-spans"
	RequirementHostComparison  ContractRequirement = "host-comparison"
	RequirementIndependentGate ContractRequirement = "independent-gate"
)

// HostingContract is the comparable contract metadata for one host stage.
type HostingContract struct {
	Stage           HostStage
	Status          ContractStatus
	SourceAuthority string
	Producer        string
	Requirements    []ContractRequirement
}
