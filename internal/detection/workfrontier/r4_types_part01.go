package workfrontier

// R4SchemaVersion is the versioned finite-frontier envelope. It is separate
// from the legacy selector schema so legacy inputs cannot silently acquire a
// root or SCC-rule interpretation.
const R4SchemaVersion = "gooo/work-frontier-r4/v1"
const (
	R4StatusPass       = "PASS"
	R4StatusBlocked    = "BLOCKED"
	R4StatusUnknown    = "UNKNOWN"
	R4StatusFailClosed = "FAIL_CLOSED"
)
const (
	R4ReasonNone                    = "NONE"
	R4ReasonRequiredInputMissing    = "REQUIRED_INPUT_MISSING"
	R4ReasonUnboundedFrontier       = "UNBOUNDED_FRONTIER"
	R4ReasonIterationExhausted      = "ITERATION_EXHAUSTED"
	R4ReasonSelectionShortfall      = "SELECTION_SHORTFALL"
	R4ReasonDuplicateSCCRule        = "DUPLICATE_SCC_RULE"
	R4ReasonConflictingSCCRule      = "CONFLICTING_SCC_RULE"
	R4ReasonMalformedGraph          = "MALFORMED_GRAPH"
	R4ReasonMalformedBinding        = "MALFORMED_BINDING"
	R4ReasonSnapshotBindingMismatch = "SNAPSHOT_BINDING_MISMATCH"
	R4ReasonPolicyBindingMismatch   = "POLICY_BINDING_MISMATCH"
	R4ReasonRegistryBindingMismatch = "REGISTRY_BINDING_MISMATCH"
)

// R4Rule is the only accepted finite rule spelling in the R4 envelope.
type R4Rule struct {
	SCCDigest      string `json:"scc_digest"`
	MaxIterations  uint64 `json:"max_iterations"`
	IterationsUsed uint64 `json:"iterations_used"`
}

// R4Input is the complete finite-frontier snapshot. RootObligationIDs and
// Rules are required envelope fields even when the graph is acyclic; an empty
// rule array is valid only when no reachable cyclic SCC exists.
type R4Input struct {
	SchemaVersion            string            `json:"schema_version"`
	SnapshotDigest           string            `json:"snapshot_digest"`
	SnapshotPayload          string            `json:"snapshot_payload"`
	PolicyDigest             string            `json:"policy_digest"`
	PolicyPayload            string            `json:"policy_payload"`
	RegistryDigest           string            `json:"registry_digest"`
	RegistryPayload          string            `json:"registry_payload"`
	MinimumSelectedPressures uint32            `json:"minimum_selected_pressures"`
	Capacity                 Capacity          `json:"capacity"`
	Pressures                []Pressure        `json:"pressures"`
	States                   []ObligationState `json:"states"`
	Paths                    []RepairPath      `json:"paths"`
	RootObligationIDs        []string          `json:"root_obligation_ids"`
	Rules                    []R4Rule          `json:"rules"`
}

// R4WorkReceipt is componentwise evidence for the bounded detector run. The
// fields are deliberately not collapsed into a weighted scalar.
type R4WorkReceipt struct {
	GraphNodes        uint64 `json:"graph_nodes"`
	GraphEdges        uint64 `json:"graph_edges"`
	ReachableNodes    uint64 `json:"reachable_nodes"`
	ReachableEdges    uint64 `json:"reachable_edges"`
	SCCs              uint64 `json:"sccs"`
	CyclicSCCs        uint64 `json:"cyclic_sccs"`
	CondensationEdges uint64 `json:"condensation_edges"`
	RuleChecks        uint64 `json:"rule_checks"`
	IterationChecks   uint64 `json:"iteration_checks"`
	PathChecks        uint64 `json:"path_checks"`
	ConflictChecks    uint64 `json:"conflict_checks"`
}
