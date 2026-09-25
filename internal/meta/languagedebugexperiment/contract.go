package languagedebugexperiment

import "fmt"

type Contract struct {
	Schema                              string `json:"schema"`
	Version                             int    `json:"version"`
	ExpectedDebugReceipts               int    `json:"expected_debug_receipts"`
	ExpectedPausedSessions              int    `json:"expected_paused_sessions"`
	ExpectedBreakpointsReached          int    `json:"expected_breakpoints_reached"`
	ExpectedTraceEvents                 int    `json:"expected_trace_events"`
	ExpectedSubjectCoherence            int    `json:"expected_subject_coherence"`
	ExpectedExecutionDigestVariants     int    `json:"expected_execution_digest_variants"`
	ExpectedCurrentEvents               int    `json:"expected_current_events"`
	ExpectedRemainingEvents             int    `json:"expected_remaining_events"`
	ExpectedGo127Runtimes               int    `json:"expected_go127_runtimes"`
	ExpectedResourceObservations        int    `json:"expected_resource_observations"`
	ExpectedUnknownBreakpointRejections int    `json:"expected_unknown_breakpoint_rejections"`
	ExpectedNonClaims                   int    `json:"expected_non_claims"`
}

func (contract Contract) Validate() error {
	if contract.Schema != "gooo/language-debug-experiment-contract/v1" || contract.Version != 1 {
		return fmt.Errorf("language debug contract identity drifted")
	}
	values := []int{contract.ExpectedDebugReceipts, contract.ExpectedPausedSessions,
		contract.ExpectedBreakpointsReached, contract.ExpectedTraceEvents,
		contract.ExpectedSubjectCoherence, contract.ExpectedExecutionDigestVariants,
		contract.ExpectedCurrentEvents, contract.ExpectedRemainingEvents,
		contract.ExpectedGo127Runtimes, contract.ExpectedResourceObservations,
		contract.ExpectedUnknownBreakpointRejections,
		contract.ExpectedNonClaims}
	for _, value := range values {
		if value < 1 {
			return fmt.Errorf("language debug contract denominator is invalid")
		}
	}
	return nil
}
