package closure

const (
	Schema                    = "gooo/metric-meta-program-ci-closure/v2"
	ProgramSchema             = "gooo/metric-meta-program/v1"
	ProgramVerificationSchema = "gooo/metric-meta-program-verification/v1"
	ExecutionPolicy           = "READ_ONLY_ARTIFACT_CLOSURE"
	ProgramExecutionPolicy    = "READ_ONLY_META_PROGRAM"
	StatusVerified            = "VERIFIED"
	WriteEffectNone           = "none"
)

func expectedActivities() []string {
	return []string{
		"ObserveCounterfactualBoundary",
		"PreserveRepositoryWorkspace",
		"ReplayCounterfactual",
		"TerminateAtFixedPoint",
	}
}
