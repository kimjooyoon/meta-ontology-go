package languagedebugexperiment

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Summary struct {
	Coordinates                 Counter  `json:"coordinates"`
	DebugReceipts               int      `json:"debug_receipts"`
	PausedSessions              int      `json:"paused_sessions"`
	BreakpointsReached          int      `json:"breakpoints_reached"`
	TraceEvents                 int      `json:"trace_events"`
	ExecutionDigestVariants     int      `json:"execution_digest_variants"`
	CurrentEvents               int      `json:"current_events"`
	RemainingEvents             int      `json:"remaining_events"`
	UnknownBreakpointRejections int      `json:"unknown_breakpoint_rejections"`
	Unknowns                    int      `json:"unknowns"`
	Compiler                    Compiler `json:"compiler"`
	Effects                     Effects  `json:"effects"`
}

type Compiler struct {
	ExecutableDigest string `json:"executable_digest"`
	Go127Runtimes    int    `json:"go127_runtimes"`
}
