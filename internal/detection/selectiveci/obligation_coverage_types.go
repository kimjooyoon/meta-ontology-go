package selectiveci

import "github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"

const ObligationCoverageSchemaVersion = "gooo/selective-ci-obligation-coverage/v1"

type CoverageDecision string

const (
	CoverageDecisionExact   CoverageDecision = "EXACT"
	CoverageDecisionUnknown CoverageDecision = "UNKNOWN"
)

type CoverageReason string

const (
	CoverageReasonComplete          CoverageReason = "COMPLETE"
	CoverageReasonNoChange          CoverageReason = "NO_CHANGE"
	CoverageReasonMissingInput      CoverageReason = "MISSING_INPUT"
	CoverageReasonInvalidInput      CoverageReason = "INVALID_INPUT"
	CoverageReasonUnsupportedSchema CoverageReason = "UNSUPPORTED_SCHEMA"
	CoverageReasonInvalidGraph      CoverageReason = "INVALID_GRAPH"
	CoverageReasonInvalidRegistry   CoverageReason = "INVALID_REGISTRY"
	CoverageReasonInvalidSnapshot   CoverageReason = "INVALID_SNAPSHOT"
	CoverageReasonStaleGraph        CoverageReason = "STALE_GRAPH"
	CoverageReasonStaleRegistry     CoverageReason = "STALE_REGISTRY"
	CoverageReasonStaleSnapshot     CoverageReason = "STALE_SNAPSHOT"
	CoverageReasonUnknownRoot       CoverageReason = "UNKNOWN_ROOT"
	CoverageReasonDuplicateRoot     CoverageReason = "DUPLICATE_ROOT"
	CoverageReasonMissingObligation CoverageReason = "MISSING_OBLIGATION"
	CoverageReasonMissingCommand    CoverageReason = "MISSING_COMMAND"
	CoverageReasonDanglingCommand   CoverageReason = "DANGLING_COMMAND"
	CoverageReasonWorkOverflow      CoverageReason = "WORK_OVERFLOW"
)

type ObligationCoverageInput struct {
	SchemaVersion  string            `json:"schema_version"`
	Graph          impactgraph.Graph `json:"graph"`
	Registry       Registry          `json:"registry"`
	SnapshotDigest string            `json:"snapshot_digest"`
	ChangedRootIDs []string          `json:"changed_root_ids"`
}

type ObligationCoverageResult struct {
	SchemaVersion             string           `json:"schema_version"`
	Decision                  CoverageDecision `json:"decision"`
	Reason                    CoverageReason   `json:"reason"`
	FullSuiteRequired         bool             `json:"full_suite_required"`
	ChangedRootCount          uint64           `json:"changed_root_count"`
	CoveredChangedRootCount   uint64           `json:"covered_changed_root_count"`
	UncoveredChangedRootCount uint64           `json:"uncovered_changed_root_count"`
	RequiredObligationCount   uint64           `json:"required_obligation_count"`
	BoundCommandCount         uint64           `json:"bound_command_count"`
	DeterministicWorkUnits    uint64           `json:"deterministic_work_units"`
	UncoveredRootIDs          []string         `json:"uncovered_root_ids"`
	RequiredObligationIDs     []string         `json:"required_obligation_ids"`
	GraphDigest               string           `json:"graph_digest"`
	RegistryDigest            string           `json:"registry_digest"`
	SnapshotDigest            string           `json:"snapshot_digest"`
	InputDigest               string           `json:"input_digest"`
	OutputDigest              string           `json:"output_digest"`
}

type CoverageInput = ObligationCoverageInput
type CoverageResult = ObligationCoverageResult
type CoverageOutput = ObligationCoverageResult
type Decision = CoverageDecision
type Reason = CoverageReason

const (
	DecisionExact   = CoverageDecisionExact
	DecisionUnknown = CoverageDecisionUnknown
	ReasonComplete  = CoverageReasonComplete
	ReasonNoChange  = CoverageReasonNoChange
	EXACT           = CoverageDecisionExact
	UNKNOWN         = CoverageDecisionUnknown
	COMPLETE        = CoverageReasonComplete
	NO_CHANGE       = CoverageReasonNoChange
)
