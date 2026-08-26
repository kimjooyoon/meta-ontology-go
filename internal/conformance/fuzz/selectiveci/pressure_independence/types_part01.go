package pressureindependence

const (
	SchemaV1       = "gooo/pressure-independence/v1"
	CorpusSchemaV1 = "gooo/pressure-independence-corpus/v1"
)

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionFailClosed Decision = "FAIL_CLOSED"
	DecisionUnknown    Decision = "UNKNOWN"
	Pass                        = DecisionPass
	FailClosed                  = DecisionFailClosed
	Unknown                     = DecisionUnknown
)

type Reason string

const (
	ReasonNone                      Reason = "NONE"
	ReasonRequiredInputMissing      Reason = "REQUIRED_INPUT_MISSING"
	ReasonInputAmbiguous            Reason = "INPUT_AMBIGUOUS"
	ReasonStaleDigest               Reason = "STALE_DIGEST"
	ReasonApplicabilityUnproven     Reason = "APPLICABILITY_UNPROVEN"
	ReasonCatalogMismatch           Reason = "CATALOG_MISMATCH"
	ReasonIndependentGroupShortfall Reason = "INDEPENDENT_GROUP_SHORTFALL"
	ReasonDuplicatePressureID       Reason = "DUPLICATE_PRESSURE_ID"
	ReasonConflictingGroupBinding   Reason = "CONFLICTING_GROUP_BINDING"
	ReasonPredicateFalse            Reason = "PREDICATE_FALSE"
	ReasonInvalidResourceReceipt    Reason = "INVALID_RESOURCE_RECEIPT"
	ReasonProvPathMissing           Reason = "PROV_PATH_MISSING"
	ReasonProvPathMalformed         Reason = "PROV_PATH_MALFORMED"
	ReasonUnboundedFrontier         Reason = "UNBOUNDED_FRONTIER"
)

// Input is the complete digest-bound fixture contract. Expected labels are
// intentionally absent; corpus expectations live outside this value.
type Input struct {
	Schema                  string           `json:"schema"`
	FixtureID               string           `json:"fixture_id"`
	AuthoritySnapshotDigest string           `json:"authority_snapshot_digest"`
	PolicyDigest            string           `json:"policy_digest"`
	RegistryDigest          string           `json:"registry_digest"`
	OracleDigest            string           `json:"oracle_digest"`
	ToolchainOptionsDigest  string           `json:"toolchain_options_digest"`
	RequestedK              uint64           `json:"requested_K"`
	MinimumIndependent      uint64           `json:"minimum_independent"`
	PressureRecords         []PressureRecord `json:"pressure_records"`
	RequiredPressureIDs     []string         `json:"required_pressure_ids"`
	GuardIDs                []string         `json:"guard_ids"`
	FinitePathIDs           []string         `json:"finite_path_ids"`
	ResourceCeilings        ResourceCeilings `json:"resource_ceilings"`

	present inputPresence
}
type inputPresence struct {
	schema, fixture, snapshot, policy, registry, oracle, toolchain bool
	requestedK, minimumIndependent, pressureRecords                bool
	requiredIDs, guardIDs, finitePaths, resourceCeilings           bool
}
type PressureRecord struct {
	PressureID          string `json:"pressure_id"`
	CategoryID          string `json:"category_id"`
	IndependenceGroupID string `json:"independence_group_id"`
	ApplicabilityRuleID string `json:"applicability_rule_id"`
}
