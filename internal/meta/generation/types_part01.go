package generation

const SchemaVersion = "gooo/self-improvement-generation/v3"

type Decision string

const DecisionPlan Decision = "PLAN"
const DecisionFixedPoint Decision = "FIXED_POINT"
const DecisionUnknown Decision = "UNKNOWN"
const DecisionRejected Decision = "REJECTED"

type Reason string

const ReasonIndependentActions Reason = "INDEPENDENT_META_OPERATIONS"
const ReasonExactFixedPoint Reason = "EXACT_FIXED_POINT"
const ReasonInvalidInput Reason = "INVALID_INPUT"
const ReasonMissingOperation Reason = "MISSING_META_OPERATION"
const ReasonApplicabilityUnproven Reason = "APPLICABILITY_UNPROVEN"
const ReasonPressureShortfall Reason = "INDEPENDENT_PRESSURE_SHORTFALL"
const ReasonFloorRegression Reason = "FLOOR_REGRESSION"

// ProofChoice makes one branch of the Munchhausen trilemma explicit.
type ProofChoice string

const ProofFoundation ProofChoice = "FOUNDATION"
const ProofCoherence ProofChoice = "COHERENCE"
const ProofRegress ProofChoice = "REGRESS"

const requestedK uint32 = 2
const minimumIndependent uint32 = 2
