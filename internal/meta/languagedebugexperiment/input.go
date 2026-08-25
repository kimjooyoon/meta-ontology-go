package languagedebugexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"

type Input struct {
	SubjectSHA        string                `json:"subject_sha"`
	ExecutableDigest  string                `json:"executable_digest"`
	Contract          Contract              `json:"contract"`
	First             languagedebug.Receipt `json:"first"`
	Second            languagedebug.Receipt `json:"second"`
	UnknownBreakpoint languagedebug.Receipt `json:"unknown_breakpoint"`
}

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
	Unknowns                    int
	RepositoryWrites            int
	MutationAuthority           bool
}
