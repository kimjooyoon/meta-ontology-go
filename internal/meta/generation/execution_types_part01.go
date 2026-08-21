package generation

const ExecutionManifestSchemaVersion = "gooo/meta-operation-execution/v2"

type ExecutionDecision string

const ExecutionDecisionProposed ExecutionDecision = "PROPOSED"
const ExecutionDecisionFixedPoint ExecutionDecision = "FIXED_POINT"
const ExecutionDecisionUnknown ExecutionDecision = "UNKNOWN"
const ExecutionDecisionRejected ExecutionDecision = "REJECTED"

type ExecutionReason string

const ExecutionReasonIndependentActions ExecutionReason = "INDEPENDENT_META_OPERATIONS"
const ExecutionReasonExactFixedPoint ExecutionReason = "EXACT_FIXED_POINT"
const ExecutionReasonInvalidPlan ExecutionReason = "INVALID_GENERATION_PLAN"
const ExecutionReasonPlanNotExecutable ExecutionReason = "PLAN_NOT_EXECUTABLE"
const ExecutionReasonPlanRejected ExecutionReason = "GENERATION_PLAN_REJECTED"

type WorkspaceMode string

const WorkspaceModeDisposable WorkspaceMode = "DISPOSABLE_WORKTREE"

type WriteBoundary string

const WriteBoundarySandboxOnly WriteBoundary = "SANDBOX_ONLY"
