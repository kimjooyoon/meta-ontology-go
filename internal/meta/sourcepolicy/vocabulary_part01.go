package sourcepolicy

type Family string
type ProofChoice string
type Relation string
type Operation string

const (
	FamilyVolume      Family = "volume"
	FamilyTopology    Family = "topology"
	FamilyDuplication Family = "duplication"
	FamilyConformance Family = "conformance"
)

// ProofChoice makes the Munchausen-trilemma branch explicit for every metric.
const (
	ProofFoundation ProofChoice = "foundation"
	ProofCoherence  ProofChoice = "coherence"
	ProofRegression ProofChoice = "regression"
)

const (
	RelationLessOrEqual Relation = "less_or_equal"
	RelationEqual       Relation = "equal"
	RelationObserve     Relation = "observe"
)

const (
	OperationObserve          Operation = "observe"
	OperationSplitGo          Operation = "split-go-declarations"
	OperationSplitGooo        Operation = "split-gooo-sections"
	OperationExtractFunction Operation = "extract-function"
	OperationPartition        Operation = "partition-directory"
	OperationSeparateKinds    Operation = "separate-directory-kinds"
	OperationModernize        Operation = "apply-go-fix"
	OperationSelectToolchain  Operation = "select-toolchain"
)
