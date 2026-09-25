package languagedebugexperiment

type facts struct {
	DebugReceipts               int
	PausedSessions              int
	BreakpointsReached          int
	TraceEvents                 int
	SubjectCoherence            int
	ExecutionDigestVariants     int
	CurrentEvents               int
	RemainingEvents             int
	Go127Runtimes               int
	UnknownBreakpointRejections int
	NonClaims                   int
	ReplayMatches               int
	ResourceObservations        int
	Unknowns                    int
	UnknownCases                []Uncertainty
	RefutedCases                []Refutation
	RepositoryWrites            int
	MutationAuthority           bool
}
