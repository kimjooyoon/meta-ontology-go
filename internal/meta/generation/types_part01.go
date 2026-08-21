package generation

const SchemaVersion = "gooo/self-improvement-generation/v1"

type Decision string

const (
	DecisionPlan       Decision = "PLAN"
	DecisionFixedPoint Decision = "FIXED_POINT"
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionRejected   Decision = "REJECTED"
)

type Reason string

const (
	ReasonIndependentActions Reason = "INDEPENDENT_META_OPERATIONS"
	ReasonExactFixedPoint     Reason = "EXACT_FIXED_POINT"
	ReasonInvalidInput        Reason = "INVALID_INPUT"
	ReasonMissingOperation    Reason = "MISSING_META_OPERATION"
	ReasonPressureShortfall   Reason = "INDEPENDENT_PRESSURE_SHORTFALL"
	ReasonFloorRegression     Reason = "FLOOR_REGRESSION"
)

// ProofChoice makes one branch of the Munchhausen trilemma explicit.
type ProofChoice string

const (
	ProofFoundation ProofChoice = "FOUNDATION"
	ProofCoherence  ProofChoice = "COHERENCE"
	ProofRegress    ProofChoice = "REGRESS"
)

const requestedK uint32 = 2
const minimumIndependent uint32 = 2
