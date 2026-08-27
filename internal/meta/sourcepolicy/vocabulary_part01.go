package sourcepolicy

type Family string
type ProofChoice string
type Relation string
type Operation string
type SubjectKind string
type Applicability string
type ApplicabilityReason string

const IndicatorSchema = "gooo/indicator-report/v3"

const (
	FamilyVolume        Family = "volume"
	FamilyTopology      Family = "topology"
	FamilyDuplication   Family = "duplication"
	FamilyRefactor      Family = "refactor"
	FamilyConformance   Family = "conformance"
	FamilyDocumentation Family = "documentation"
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
	SubjectKindProjectRoot    SubjectKind = "PROJECT_ROOT"
	SubjectKindDirectory      SubjectKind = "DIRECTORY"
	SubjectKindFile           SubjectKind = "FILE"
	SubjectKindFunction       SubjectKind = "FUNCTION"
	SubjectKindSourceFragment SubjectKind = "SOURCE_FRAGMENT"
	SubjectKindRepository     SubjectKind = "REPOSITORY"
)

const (
	ApplicabilityApplicable    Applicability = "APPLICABLE"
	ApplicabilityNotApplicable Applicability = "NOT_APPLICABLE"

	ApplicabilityReasonCatalogApplicable  ApplicabilityReason = "CATALOG_APPLICABLE"
	ApplicabilityReasonRootTopologyExempt ApplicabilityReason = "ROOT_TOPOLOGY_EXEMPT"
	ApplicabilityReasonRootREADMEExempt   ApplicabilityReason = "ROOT_README_EXEMPT"
	ApplicabilityReasonWorkflowDiscovery  ApplicabilityReason = "GITHUB_WORKFLOW_DISCOVERY_ROOT"

	WorkflowDiscoveryObservationDetail = "topology=github-workflow-discovery-root"
)

const (
	OperationObserve           Operation = "observe"
	OperationSplitGo           Operation = "split-go-declarations"
	OperationSplitGooo         Operation = "split-gooo-sections"
	OperationExtractFunction   Operation = "extract-function"
	OperationInspectWrapper    Operation = "inspect-wrapper"
	OperationCollapseAssign    Operation = "collapse-assign-return"
	OperationPartition         Operation = "partition-directory"
	OperationSeparateKinds     Operation = "separate-directory-kinds"
	OperationExemptRoot        Operation = "exempt-project-root-topology"
	OperationExemptRootREADME  Operation = "exempt-project-root-readme"
	OperationRequireRootREADME Operation = "require-project-root-readme"
	OperationPreserveWorkflow  Operation = "preserve-workflow-discovery"
	OperationModernize         Operation = "apply-go-fix"
	OperationSelectToolchain   Operation = "select-toolchain"
	OperationMeasureProgress   Operation = "measure-integration-progress"
)
